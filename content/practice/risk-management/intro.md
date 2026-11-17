**The prompt:** *"Design a risk management system for our payments platform. Every transaction should be checked before we authorise it."*

This comes up a lot at fintech and payments companies (it may be called a *fraud engine*, *risk engine* or *rule engine*). At its core it's a **rule engine**, and the LLD signals are:

- rules as small, pluggable **strategies** behind one interface
- a clear separation between **rules** (what's suspicious), the **policy** (what we do about it) and the **actions** (audit, review queue, alerts)
- **stateful** checks (velocity) done correctly under concurrency
- **runtime-configurable** rules without restarts or races

| Phase | Interview time |
|---|---|
| Clarify requirements | 5 min |
| Entities, class diagram | 8–10 min |
| Evaluation and decision | 7 min |
| Velocity, concurrency, config | 8 min |
| Code | 15 min |
| Follow-ups | 5 min |
