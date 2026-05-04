---
title: "Slack thread C0AKH5SNGKH 1774051400.233059"
wiki_page_source: "slack_message"
source_document_id: "srcdoc_b36a571fb891dc57a0ee8a6bffe5e227"
source_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1774051400.233059"
source_session_key: "slack:T045QQQQ7CZ:C0AKH5SNGKH:1774051400.233059"
source_revision_ids:
  - "srcrev_24ecc1b3a8183b412448c4dfe655c15e"
  - "srcrev_388a789e3fac3f25a6be255ed7fbb056"
  - "srcrev_3c271c8145125dba9f0064e04aa41567"
  - "srcrev_5b1046d2dce3e4c1875b55ff27fc08a3"
  - "srcrev_8ad0d4e4662d3ef126efd9f58e72ed04"
  - "srcrev_95a1b6265dd971cd25a7bb75f6fd873d"
  - "srcrev_9a47c08cbfff6e0006ecbd19bd79adfd"
  - "srcrev_bd39a52eadd29e815240856ea8d403a5"
  - "srcrev_c4ce280996b870a9223257452ba81fa3"
  - "srcrev_d468065163a96885e3fb4fb7c63d279c"
  - "srcrev_d9a9ad0651b3a6802a387931adb8cabc"
  - "srcrev_dd010107d802d0042a83c6fce7ac8683"
  - "srcrev_dea7a504fe1d3ae2d47b58da77f1e1ea"
  - "srcrev_fe3ccf1abfd690a29d657ddac050b0e8"
conflicts: []
---

# Slack thread C0AKH5SNGKH 1774051400.233059

## Compiled Evidence

### Citation 1

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_fe3ccf1abfd690a29d657ddac050b0e8`
- `chunk_id`: `srcchunk_2e3b61c48cb7ec3dbe117db994aa4b7a`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774051400.233059`

<@U04L0DD6B6F> <@U08HVGL6LDR> <@U0772SH7BRA> I created a github action from my jobs scraper to get additional data in the csv file and delete removed / old jobs
The action runs every 3 days and pushes the updated csv on the `scraping-actions` branch
<https://github.com/piplabs/jobs-app-scraper>

### Citation 2

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_9a47c08cbfff6e0006ecbd19bd79adfd`
- `chunk_id`: `srcchunk_6f81eb540667cf2ef94e1b14e8526c4a`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774051440.506719`

this is amazing!

### Citation 3

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_c4ce280996b870a9223257452ba81fa3`
- `chunk_id`: `srcchunk_fa76886c1a008e6f787ffb2b021b35f1`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774051457.870829`

are we filtering for specific jobs in regions or how does the high level logic work

### Citation 4

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_8ad0d4e4662d3ef126efd9f58e72ed04`
- `chunk_id`: `srcchunk_1bc0294e67ad15cfab23b67354d6da19`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774051484.121649`

Sick

### Citation 5

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_95a1b6265dd971cd25a7bb75f6fd873d`
- `chunk_id`: `srcchunk_1269e639a219008d37dc2b4fb1297231`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774051605.120679`

It simply scraps the Linkedin public API using keywords matching the job agents fields:
```KEYWORDS = [
    # general-software-engineer
    "Software Engineer",
    "Software Developer",
    "Backend Engineer",
    # product-manager
    "Product Manager",
    "Technical Product Manager",
    # product-designer
    "Product Designer",
    "UX Designer",
    # data-analyst
    "Data Analyst",
    "Business Intelligence Analyst",
    # data-engineer
    "Data Engineer",
    "Analytics Engineer",
    # ai-ml-engineer
    "Machine Learning Engineer",
    "AI Engineer",
    # technical-recruiter
    "Technical Recruiter",
    "Engineering Recruiter",
    # solutions-engineer
    "Solutions Engineer",
    "Sales Engineer",
    # customer-success-manager
    "Customer Success Manager",
    "Client Success Manager",
    # product-marketing-manager
    "Product Marketing Manager",
    "Growth Marketing Manager",
]```
No specific region or other filters yet but we could definitely add them

### Citation 6

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_388a789e3fac3f25a6be255ed7fbb056`
- `chunk_id`: `srcchunk_99e99acac410aa8658e7e1f64bf64ab8`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774052451.545919`

interesting - mind sharing some architecutre on the scraping? is it thru puppeteer or directly thru calling their apis? cause i assume they have some rate limits around that

also have you already handle how to deal with dedupes?

### Citation 7

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_d9a9ad0651b3a6802a387931adb8cabc`
- `chunk_id`: `srcchunk_c27970ab7b1af93307da7f83e4b8d8f7`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774198816.516919`

It simply uses the likedin public API with random sleep times between requests and backoffs when encountering a 429!

After fetching the job urls from linkedin pages, it uses 3 threads to scrap the descriptions and remove dead jobs. The script adds the data in the existing csv file, and handles dedupes and jobs that are too old using pandas.

### Citation 8

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_bd39a52eadd29e815240856ea8d403a5`
- `chunk_id`: `srcchunk_1d7b5eb74daf9688a5a6876ec33c5a9e`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774199185.914119`

Usually how many pages are there for different roles

### Citation 9

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_24ecc1b3a8183b412448c4dfe655c15e`
- `chunk_id`: `srcchunk_477d782685d0894de0355c711788510c`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774200205.439829`

On each run, the script fetches 4 pages per keyword, each page containing 10 jobs =&gt; 80 new jobs per job category before removing duplicates. Currently the csv file stored in the repo contains 100 to 130 jobs per agent

### Citation 10

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_dea7a504fe1d3ae2d47b58da77f1e1ea`
- `chunk_id`: `srcchunk_823610515270deb1e56ad141bf250e34`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774200318.097349`

Great work! Curious how varied are the locations? Because each job / role also is strongly tied to a specific location. I suspect in the US it's mostly SF or NY. But we need to try to cover more and potentially see if this works globally

### Citation 11

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_5b1046d2dce3e4c1875b55ff27fc08a3`
- `chunk_id`: `srcchunk_d7358901e9ee324db8aeab2d7709e805`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774242154.494999`

Just had Claude producing an analytics script and the result for cities looks like:

```  New York, NY: 232
  United States: 213
  South Bay / Peninsula, CA: 118
  San Francisco, CA: 105
  Los Angeles, CA: 54
  Seattle, WA: 40
  Chicago, IL: 28
  Austin, TX: 23
  Boston, MA: 23
  Washington, DC: 20
  Miami, FL: 18
  Atlanta, GA: 17
  Dallas, TX: 17
  Philadelphia, PA: 14
  Houston, TX: 13```
Then a long tail of small cities

### Citation 12

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_dd010107d802d0042a83c6fce7ac8683`
- `chunk_id`: `srcchunk_a82e09b70698a2ef7f54716cd45aaed3`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774244552.861629`

i see this is useful!

### Citation 13

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_d468065163a96885e3fb4fb7c63d279c`
- `chunk_id`: `srcchunk_8def4bfbb590f27baeb5b3b92aa27f7b`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774244568.863749`

is there an api to access some of this?

### Citation 14

- `source_document_id`: `srcdoc_b36a571fb891dc57a0ee8a6bffe5e227`
- `source_revision_id`: `srcrev_3c271c8145125dba9f0064e04aa41567`
- `chunk_id`: `srcchunk_a54ab65abbd7f3cadacb96d38be49d9e`
- `native_locator`: `slack:C0AKH5SNGKH:1774051400.233059:1774281751.588109`

Do you mean the analytics or the data itself?

