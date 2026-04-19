import { describe, expect, it } from "vitest";

import { slackMrkdwnToSegments } from "./formatted-message";

describe("slackMrkdwnToSegments", () => {
  it("uses resolved Slack user and channel names from metadata when mrkdwn omits labels", () => {
    expect(
      slackMrkdwnToSegments(
        "Ping <@U0ASDQKU3UL> in <#C0AKH5SNGKH> &amp; review <https://example.com/runbook|runbook> <!here>",
        {
          userNames: { U0ASDQKU3UL: "blake" },
          channelNames: { C0AKH5SNGKH: "ops-incidents" }
        }
      )
    ).toEqual([
      { text: "Ping @blake in #ops-incidents & review " },
      { text: "runbook", href: "https://example.com/runbook" },
      { text: " @here" }
    ]);
  });

  it("falls back to raw Slack IDs when no display labels are available", () => {
    expect(slackMrkdwnToSegments("Ping <@U0ASDQKU3UL> in <#C0AKH5SNGKH>")).toEqual([{ text: "Ping @U0ASDQKU3UL in #C0AKH5SNGKH" }]);
  });

  it("uses resolved Slack user and channel names for bare Slack IDs in plain text", () => {
    expect(
      slackMrkdwnToSegments(
        "Hello @U0ASDQKU3UL, please check #C0AKH5SNGKH and #C0AL7EKNHDF.",
        {
          userNames: { U0ASDQKU3UL: "blake" },
          channelNames: {
            C0AKH5SNGKH: "depin-backend",
            C0AL7EKNHDF: "numo-project"
          }
        }
      )
    ).toEqual([
      { text: "Hello @blake, please check #depin-backend and #numo-project." }
    ]);
  });
});
