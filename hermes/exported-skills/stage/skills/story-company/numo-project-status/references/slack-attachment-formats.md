# Slack Attachment Format Convention

**Rule:** When delivering visual artifacts (reports, charts, diagrams, dashboards) to Slack, attach them as **PNG or SVG image files** — never raw HTML.

**Why:** Slack threads do not render inline HTML. Raw `.html` files appear as unreadable plaintext or an undisplayable attachment. Image files (PNG/SVG) render natively with a preview in Slack threads.

## Conversion Strategy

### For HTML reports/diagrams (e.g., architecture-diagram output, audio_quality_report.html):

1. Generate the HTML file as usual
2. Convert to PNG using one of:
   - **Headless browser screenshot** (preferred for fidelity):
     ```bash
     # Using chromium/chrome headless
     chromium --headless --disable-gpu --screenshot=/path/to/output.png \
       --window-size=1280,900 --virtual-time-budget=3000 \
       file:///path/to/report.html
     ```
   - **wkhtmltoimage** (lighter but less CSS support):
     ```bash
     wkhtmltoimage --width 1280 /path/to/report.html /path/to/output.png
     ```
3. Attach the PNG (not the HTML) to the Slack message using `MEDIA:/path/to/output.png` in the message text
4. Keep the HTML as an artifact for reference — but the Slack attachment should be the image

### For SVG diagrams:
- SVG files can be attached directly to Slack (`MEDIA:/path/to/diagram.svg`) — Slack renders SVG with good fidelity
- For complex SVGs with web fonts, consider PNG conversion for consistency

### For tables/data grids:
- Screenshot rendered tables as PNG
- Or format data directly as Slack markdown tables in the message body

## Slack MEDIA attachment syntax

```
MEDIA:/absolute/path/to/file.png
```

Include this token in the message body passed to `send_message`. The platform delivers it as a native attachment.

## Detection: when to apply this

- Any time `write_file` produces a `.html` file intended for Slack delivery
- Any `architecture-diagram` invocation where the user is in Slack
- Any `numo-project-status` report that includes a "full visual report attached"
- Any other creative skill output (charts, infographics, dashboards) destined for Slack

## Related

- `architecture-diagram` skill: also produces HTML and should follow this convention when delivering to Slack
- `send_message` documentation: MEDIA token format
