import { describe, expect, it } from "vitest";

import { slackMrkdwnToSegments } from "./formatted-message";

describe("slackMrkdwnToSegments", () => {
  it("normalizes Slack mentions, channels, links, and escapes", () => {
    expect(
      slackMrkdwnToSegments("Ping <@U0ASDQKU3UL> in <#C0AKH5SNGKH|ops-incidents> &amp; review <https://example.com/runbook|runbook> <!here>")
    ).toEqual([
      { text: "Ping " },
      { text: "@U0ASDQKU3UL" },
      { text: " in " },
      { text: "#ops-incidents" },
      { text: " & review " },
      { text: "runbook", href: "https://example.com/runbook" },
      { text: " " },
      { text: "@here" }
    ]);
  });
});
