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
      { text: "Ping " },
      { text: "@blake" },
      { text: " in " },
      { text: "#ops-incidents" },
      { text: " & review " },
      { text: "runbook", href: "https://example.com/runbook" },
      { text: " " },
      { text: "@here" }
    ]);
  });

  it("falls back to raw Slack IDs when no display labels are available", () => {
    expect(slackMrkdwnToSegments("Ping <@U0ASDQKU3UL> in <#C0AKH5SNGKH>")).toEqual([
      { text: "Ping " },
      { text: "@U0ASDQKU3UL" },
      { text: " in " },
      { text: "#C0AKH5SNGKH" }
    ]);
  });
});
