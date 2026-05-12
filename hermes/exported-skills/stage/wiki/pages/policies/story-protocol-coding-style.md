---
title: "Story Protocol Coding Style"
type: "policy"
slug: "policies/story-protocol-coding-style"
freshness: "2023-10-06T20:07:00Z"
tags:
  - "coding-style"
  - "conventions"
  - "solidity"
owners: []
source_revision_ids:
  - "srcrev_58c4c9eebab4810f942d6570d946acb8"
conflict_state: "none"
---

# Story Protocol Coding Style

## Summary

A compilation of coding style conventions for Story Protocol's Solidity codebase, covering formatting, naming, imports, and testing.

## Claims

- Use `///` for all NatSpec comments. `claim:claim_1_1` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Imports must be ordered with external imports first, then alphabetical. Only explicit imports are allowed. `claim:claim_1_2` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- All interfaces must be placed inside an `/interface` folder, mirroring the folder structure of the implementations. `claim:claim_1_3` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Acronyms in CamelCase must be all uppercase (e.g., URI, IPAsset). For variables, use lowercase prefix (e.g., ipAsset) and uppercase suffix (e.g., tokenURI). `claim:claim_1_4` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Interfaces contain methods and events. Structs that appear in an interface must be placed in a library, with a single library per module. Errors follow the pattern from Sablier's Errors.sol. `claim:claim_1_5` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Test names must follow the pattern `test_contextCamel_descriptionOfTheTestCamel`, where context is the method, contract, or functionality. `claim:claim_1_6` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Functions within a contract must be ordered by functionality (delimited by comment blocks), then by visibility: external, public, internal, private, and within each visibility by state mutability: write, view, pure. `claim:claim_1_7` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Maximum line length is 120 characters. `claim:claim_1_8` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Solhint must be configured with the `no-unused-import` rule set to error. `claim:claim_1_9` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Function parameter names must use a trailing underscore suffix. `claim:claim_1_10` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- If a function call exceeds 120 characters, each argument must be placed on a separate line. `claim:claim_1_11` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Long if statements (>120 characters) must break conditions onto separate lines, with logical operators at the end of lines. Nested conditions should be indented. `claim:claim_1_12` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Long assignments (>120 characters) must break after the assignment operator, with the right-hand side indented on the next line. `claim:claim_1_13` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- The general order of functions should follow the official Solidity style guide. `claim:claim_1_14` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`
- Naming conventions: Contracts use CamelCase (adjectiveNoun), structs use nouns, events use past-tense verbs, functions use verbNoun, local variables use nouns/compound nouns, booleans use `is` prefix (e.g., isValid), modifiers use prepositionNoun (e.g., onlyOwner), and errors follow a separate pattern. `claim:claim_1_15` `confidence:1.00`
  - citation: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720) `source_document_id=srcdoc_801b40fb89e21591748efcff511a2e2a` `source_revision_id=srcrev_58c4c9eebab4810f942d6570d946acb8` `chunk_id=srcchunk_f2b827ee6145ad39c279383574d8f2cf` `native_locator=https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720` `source_timestamp=2023-10-06T20:07:00Z`

## Sources

- `source_document_id`: `srcdoc_801b40fb89e21591748efcff511a2e2a`
- `source_revision_id`: `srcrev_58c4c9eebab4810f942d6570d946acb8`
- `source_url`: [Notion source](https://www.notion.so/Story-Protocol-Coding-Style-0bb1fc9caa784b738aae507d305b8720)
