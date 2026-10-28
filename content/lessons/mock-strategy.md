---
title: How to Run a Mock Interview on Yourself
summary: The protocol for the 4 timed mocks in days 13–14, so they're real practice and not just more exercises.
---

The mocks only work if they're **uncomfortable in the same way the real interview is**: a clock, no hints, and talking while thinking.

## Before you press Start

- Close everything else. Phone in another room.
- Pick your interview language in the top bar. It locks for the attempt.
- Have a scratch notepad (paper is best) for the diagram.
- **Decide your opening lines** so you don't freeze: *"Let me make sure I understand the scope. I'll list the use cases I plan to support, and you tell me what I've missed."*

## During the 60 minutes

| Minute | Do this |
|---|---|
| 0–6 | Read the prompt. Write **numbered requirements** and **assumptions** as comments at the top of the file. Say them out loud |
| 6–12 | Entities, enums and the main service's method signatures, as code skeletons rather than a pretty diagram |
| 12–40 | Core flow top-down. Press **Run** early and often, even with stubs |
| 40–50 | Edge cases and errors. Then concurrency: name the shared state and add the lock |
| 50–55 | Press **Run tests**. Fix what's cheap to fix |
| 55–60 | Write a `// Extensions:` comment listing how you'd handle 2–3 likely follow-ups |

**Talk out loud the whole time.** It feels silly alone, but the round is graded on communication, and narrating while you type is a skill you have to train.

## After you press Finish

1. Hidden tests run automatically. Don't treat them as the grade. They're one input to the **"working code"** criterion.
2. Score yourself on the rubric. **Be strict.** "Strong" means you'd bet on a hire signal for that line.
3. Write one sentence in "notes for next time", the single most important fix.
4. Open the problem as a normal exercise, read the reference solution, and compare structure, not just correctness.
5. Re-read *20 Mistakes That Cost Offers* and tick off the ones you made.

## Reading your score

| Score | Meaning |
|---|---|
| 85%+ | Interview-ready on this problem type. Move on |
| 70–84% | Borderline. Redo the weakest rubric line on a different problem |
| < 70% | Redo the same mock in 24 h in the *other* language, or re-read the matching walkthrough |

## If the real interview format differs

- **45-minute OOD at FAANG:** compress the code phase. Write the core classes and the main method fully, and stub the rest with clear names.
- **90–120-minute machine coding:** expand everything. Add a `main` demo and simple unit tests, and expect the interviewer to run your code and to add a requirement halfway through.
