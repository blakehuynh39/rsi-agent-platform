import type { ReactNode } from "react";
import type { JsonObject } from "@/types";

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

function decodeSlackEntities(text: string) {
  return text.replaceAll("&amp;", "&").replaceAll("&lt;", "<").replaceAll("&gt;", ">");
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

function stringRecordFromMetadataValue(value: unknown): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  const out: Record<string, string> = {};
  for (const [key, item] of Object.entries(value)) {
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

  return segments;
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
