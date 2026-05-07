---
name: state-machine-design-review
description: "Analyze flat-enum state machines for orthogonal-dimension conflation — identify overloaded enums, map bugs to conflation, and decide when to decompose into product types."
version: 1.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [state-machine, enum-design, Rust, architecture-review, product-type]
    related_skills: [systematic-debugging, writing-plans]
prerequisites:
  commands: [grep, sed, git]
---

# State Machine Design Review

Analyze whether a finite-state machine's flat enum design is appropriate or "overloaded" — conflating multiple orthogonal dimensions into a single set of variants. Determine whether to stay flat, add missing states, or refactor into a product type (multiple enums).

## When to Use

- User asks whether a state machine enum is "overloaded" or conflating concerns
- User asks "should we have multiple enums instead of one?"
- User reports bugs in a state machine and suspects the enum design is the root cause
- User is planning to add states and wants guidance on whether the flat approach still scales
- Pre-PR design review of a state machine addition

## 1. Gather Evidence

Read the actual code — do not speculate from memory:

```bash
# Find the enum definition
grep -rn "enum.*State" <repo>/src/

# Find the transition function (reducer)
grep -rn "fn reduce\|fn transition\|fn apply" <repo>/src/

# Find all callers that feed events into the reducer
grep -rn "reduce_\|Event::\|Transition::" <repo>/src/
```

Read design documents and bug-tracker history to understand the evolution of the state machine. Design plans that document _why_ states were added reveal the dimensions that got folded in.

## 2. Identify Orthogonal Dimensions

For each variant in the flat enum, ask: **what independent fact does this variant communicate?**

Group variants by the dimensions they answer:

| Dimension | Example question the dimension answers |
|---|---|
| Nonce/ordering bookkeeping | "Has the DB advanced past this nonce or not?" |
| Broadcast outcome | "Did the broadcast succeed, fail, or is it uncertain?" |
| Receipt/confirmation | "Did the chain accept or reject this tx?" |
| Replacement tracking | "Is this attempt being replaced by a bump?" |
| Termination status | "Is this state final (no further transitions)?" |

A dimension is **orthogonal** if a variant from one dimension can combine with variants from another dimension to produce valid states. If two dimensions always change together (every transition flips both simultaneously), they may be a single dimension in disguise.

**Red flag:** A single variant means two different things depending on external context. Example: `AwaitingVerify` means both "nonce NOT consumed, should retry" _and_ "nonce WAS consumed, should recover hash" — the confirmer has to call `chain.pending_nonce()` to distinguish them at runtime.

## 3. Map Concrete Bugs to Dimensional Conflation

This is the **strongest evidence** for overloading. For each bug observed in production:

| Bug | Overloaded variant | What was conflated | Fix |
|---|---|---|---|
| Nonce drift alert | `AwaitingVerify` | "nonce unused" vs "nonce consumed" | Confirmer re-derives from chain query at runtime |
| Revert cascade (11 jobs) | `AwaitingVerify` | "nothing broadcast (estimate reverted)" vs "broadcast uncertain (RPC timeout)" | Added `Discarded` state |
| Ordering reversal blocked | n/a (state missing) | No state for "nonce committed, not broadcast" | Needs new `NonceCommitted` state |

If bugs required adding a _new_ flat variant for what should have been a value of an existing dimension, that's evidence the flat enum is hitting its ceiling.

## 4. Evaluate Flat vs Product-Type

**Flat enum** (current approach):
- One `enum` with N variants representing valid combinations
- Reducer is one flat `match (state, event)` — simple and auditable
- Adding a state = adding one variant + one reducer arm
- Works well up to ~12-15 variants

**Product type** (decomposition):
- Multiple small enums, one per dimension
- Reducer input is a struct of these enums
- Invalid combinations are **type-level impossible** — no runtime checks needed
- Adding a dimension = one new enum + update struct, no multiplication
- More boilerplate; may need a transition table or nested matches

**Decision threshold: ~15 flat variants.** Below this, the flat enum is simpler and easier to audit. Above this, the dimensionality cost overtakes the simplicity benefit. Also refactor if:
- A 5th orthogonal dimension is being added (the complexity compounds)
- Multiple bugs in a short period trace back to dimension conflation
- New team members cannot predict transitions from the state diagram alone

## 5. Make the Recommendation

Structure the answer as:

1. **Short answer:** yes/no on overloading, plus the immediate action
2. **Current dimensions:** table of what the enum really represents
3. **Bug evidence:** concrete bugs traced to conflation
4. **Alternative design:** what a product type would look like
5. **Recommendation:** immediate action (usually: add missing state) + long-term trigger (when to refactor)
6. **Summary table:** quick-reference Q&A format

### Recommendation template

```
Short answer: [Mildly / Significantly] overloaded — [N] dimensions in [M] flat variants.
Immediate action: [Add state X / Do nothing / Full refactor].
Long-term: revisit when [threshold condition].
```

## 6. Deliver

For Slack-based workflows, post the analysis to the ingress thread. Keep the most actionable parts (recommendation + summary table) accessible without scrolling. For longer analyses, use a summary table at both the top and bottom.

### Pitfalls

1. **Don't refactor just because the enum is flat.** Flat enums exist for a reason — they're simple, auditable, and well-supported by pattern matching. Only push for decomposition when you can map _concrete bugs_ to dimensional conflation.
2. **Don't inline the full code in the analysis.** Link to files at specific line numbers. The analysis should be readable without code literacy.
3. **The "~15 variant threshold" is heuristic**, not a law. Calibrate against team size, codebase maturity, and bug frequency.
4. **Legacy variants (like `Sending`) inflate the count.** Count _active_ variants separately from deprecated ones when evaluating the threshold.
5. **Be concrete about the alternative.** Don't just say "split the enum" — show the proposed product-type enums and explain what invalid combinations become impossible.

### References

- `references/depin-backend-txattempt-enum-analysis.md` — Full session analysis of depin-backend's `TxAttemptState` enum with 10 variants, 4 dimensions, and 3 concrete bugs.
