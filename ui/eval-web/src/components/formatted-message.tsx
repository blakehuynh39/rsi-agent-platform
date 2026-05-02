import type { ReactNode } from "react";
import type { JsonObject, JsonValue } from "@/types";

type TextSegment = {
  text: string;
  href?: string;
};

type SlackEntityLabels = {
  userNames: Record<string, string>;
  channelNames: Record<string, string>;
};

const emptySlackEntityLabels: SlackEntityLabels = {
  userNames: {},
  channelNames: {}
};

const plainSlackEntityPattern = /(^|[^A-Za-z0-9])([@#])([A-Z0-9]{8,})(?=$|[^A-Za-z0-9])/g;

function decodeSlackEntities(text: string) {
  return text.replace(/&amp;/g, "&").replace(/&lt;/g, "<").replace(/&gt;/g, ">");
}

function splitSlackToken(token: string): [value: string, label?: string] {
  const separatorIndex = token.indexOf("|");
  if (separatorIndex === -1) return [token];
  return [token.slice(0, separatorIndex), token.slice(separatorIndex + 1)];
}

function isLinkTarget(value: string) {
  return /^(https?:\/\/|mailto:)/i.test(value);
}

function parseSlackToken(token: string, labels: SlackEntityLabels): TextSegment {
  const [value, label] = splitSlackToken(token);

  if (value.startsWith("@")) {
    const id = value.slice(1);
    return { text: `@${label || labels.userNames[id] || id}` };
  }

  if (value.startsWith("#")) {
    const id = value.slice(1);
    return { text: `#${label || labels.channelNames[id] || id}` };
  }

  if (value.startsWith("!subteam^")) {
    return { text: label || "@group" };
  }

  if (value.startsWith("!date^")) {
    return { text: label || value.slice("!date^".length) };
  }

  if (value.startsWith("!")) {
    return { text: `@${label || value.slice(1).split("^")[0]}` };
  }

  if (isLinkTarget(value)) {
    return {
      text: label || (value.startsWith("mailto:") ? value.slice("mailto:".length) : value),
      href: value
    };
  }

  if (label) {
    return { text: label };
  }

  return { text: value };
}

function stringRecordFromMetadataValue(value: JsonValue | undefined): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  const out: Record<string, string> = {};
  for (const key of Object.keys(value)) {
    const item = value[key];
    if (typeof item !== "string") continue;
    const trimmedKey = key.trim();
    const trimmedValue = item.trim();
    if (!trimmedKey || !trimmedValue) continue;
    out[trimmedKey] = trimmedValue;
  }
  return out;
}

function slackEntityLabelsFromMetadata(metadata?: JsonObject): SlackEntityLabels {
  if (!metadata) return emptySlackEntityLabels;
  return {
    userNames: stringRecordFromMetadataValue(metadata.slack_user_names),
    channelNames: stringRecordFromMetadataValue(metadata.slack_channel_names)
  };
}

function plainSlackEntitySegment(prefix: "@" | "#", id: string, labels: SlackEntityLabels): TextSegment {
  if (prefix === "@") {
    return { text: `@${labels.userNames[id] || id}` };
  }
  return { text: `#${labels.channelNames[id] || id}` };
}

function expandPlainSlackEntitySegments(segments: TextSegment[], labels: SlackEntityLabels): TextSegment[] {
  const out: TextSegment[] = [];

  for (const segment of segments) {
    if (segment.href || !segment.text) {
      out.push(segment);
      continue;
    }

    let cursor = 0;
    plainSlackEntityPattern.lastIndex = 0;
    let match: RegExpExecArray | null;

    while ((match = plainSlackEntityPattern.exec(segment.text)) !== null) {
      const boundary = match[1] || "";
      const prefix = match[2] as "@" | "#";
      const id = match[3];
      const tokenStart = match.index + boundary.length;
      const tokenEnd = tokenStart + prefix.length + id.length;

      if (match.index > cursor) {
        out.push({ text: segment.text.slice(cursor, match.index) });
      }
      if (boundary) {
        out.push({ text: boundary });
      }
      out.push(plainSlackEntitySegment(prefix, id, labels));
      cursor = tokenEnd;
    }

    if (cursor < segment.text.length) {
      out.push({ text: segment.text.slice(cursor) });
    }
  }

  return out;
}

function mergeAdjacentTextSegments(segments: TextSegment[]): TextSegment[] {
  const out: TextSegment[] = [];
  for (const segment of segments) {
    const previous = out[out.length - 1];
    if (!segment.href && previous && !previous.href) {
      previous.text += segment.text;
      continue;
    }
    out.push({ ...segment });
  }
  return out;
}

export function slackMrkdwnToSegments(text: string, labels: SlackEntityLabels = emptySlackEntityLabels): TextSegment[] {
  const decoded = decodeSlackEntities(text);
  const matcher = /<([^>\n]+)>/g;
  const segments: TextSegment[] = [];
  let cursor = 0;
  let match: RegExpExecArray | null;

  while ((match = matcher.exec(decoded)) !== null) {
    if (match.index > cursor) {
      segments.push({ text: decoded.slice(cursor, match.index) });
    }
    segments.push(parseSlackToken(match[1], labels));
    cursor = match.index + match[0].length;
  }

  if (cursor < decoded.length) {
    segments.push({ text: decoded.slice(cursor) });
  }

  return mergeAdjacentTextSegments(expandPlainSlackEntitySegments(segments, labels));
}

export function FormattedMessage(props: { source?: string; text: string; metadata?: JsonObject }) {
  const segments = props.source === "slack"
    ? slackMrkdwnToSegments(props.text, slackEntityLabelsFromMetadata(props.metadata))
    : [{ text: props.text }];

  return (
    <span className="message-text">
      {segments.map((segment, index) => {
        const key = `${segment.href || "text"}-${index}`;
        if (segment.href) {
          return (
            <a key={key} className="detail-link" href={segment.href} target="_blank" rel="noreferrer">
              {segment.text}
            </a>
          );
        }
        return <MessageTextSegment key={key}>{segment.text}</MessageTextSegment>;
      })}
    </span>
  );
}

function MessageTextSegment(props: { children: ReactNode }) {
  return <>{props.children}</>;
}
