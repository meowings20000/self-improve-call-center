# LingoLift 3–5 Minute Demo Script

## 0:00–0:30 — Frame the product

Open `http://localhost:3000`.

> “LingoLift is a small adaptive English practice app that works fully offline after setup—no login, tracking, AI key, or external model. The loop is prompt, learner action, immediate feedback, then retry or a targeted next item.”

Point out the daily XP goal, streak, hearts, lesson progress, and focus-skill label.

## 0:30–1:25 — Show a correct attempt

Choose **She walks to work every day** and select **Check answer**.

> “The learner gets immediate correctness, 10 XP, and the exact grammar rule—not praise alone. The answer is not sent with the question; it is revealed by the server only after an attempt.”

Select **Continue** and note that the progress bar and XP changed.

## 1:25–2:25 — Show error, retry, and adaptation

On the new item, deliberately choose a wrong answer and check it.

> “An incorrect response costs a heart and lowers only this skill signal. The explanation tells the learner what pattern to use. Try again keeps the same item so the learner can correct it immediately.”

Select **Try again**, choose the correct answer, and check. If demonstrating targeted routing instead, choose **Continue** directly after a wrong response and point out that the next unseen question has the same focus skill.

## 2:25–3:15 — Explain the skill map

Open **Skill map**.

> “This dashboard does not pretend to be a proficiency test. It shows transparent formative signals: sampled accuracy, practice scores, and the weakest skill. A negative skill queues another authored item in that area.”

Show “Needs focus” and the next-focus explanation.

## 3:15–4:10 — Learning basis and implementation

Open `TECHNICAL_NOTE.md` briefly.

> “The learning interaction maps practice testing to repeated retrieval, drawing on Dunlosky and colleagues’ 2013 review. The feedback is task-specific and action-oriented, informed by Hattie and Timperley’s 2007 feedback model. The implementation is a dependency-free browser client and a tested Go domain/API with 16 authored exercises.”

## 4:10–4:40 — Privacy and limits

> “Answers remain in backend memory and disappear on restart. There are no third-party runtime requests. This is a transparent proof of concept, not CEFR placement or a validated score; production use needs expert content review, more items, accessibility testing, and consent-based persistence controls.”

End on the lesson screen and invite one more question.
