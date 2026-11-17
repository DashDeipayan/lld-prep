**The prompt:** *"Design a notification system. Other services should be able to notify users by email, SMS, push or in-app."*

This is one of the most common LLD prompts, and it sits on the border with system design. In an LLD round, the interviewer wants to see:

- **channels behind one interface**, so adding WhatsApp is a new class (Strategy/Adapter)
- **user preferences and filters** as composable checks (Chain of Responsibility)
- **templates** separated from delivery
- **reliability**: async queues, retries with backoff, idempotency, rate limits and delivery status

| Phase | Interview time |
|---|---|
| Clarify requirements | 3–5 min |
| Entities, class diagram | 8–10 min |
| Send flow | 5 min |
| Patterns | 5 min |
| Reliability and concurrency | 7 min |
| Code | 15 min |
| Follow-ups | 5 min |
