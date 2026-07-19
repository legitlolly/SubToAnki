# SubToAnki
An application for quickly generating anki flashcards from subtitles.


Current State - 
1. Able to ingest JMDict_e.gz into local lookup.db.
2. Lookup a given word in kana/kanji form and display all possible readings.

TODO:
1. Add priority sorting so that words show up in order of commonality.
2. Create api for lookup feature to eventually be hosted on my local server.
3. Integrate card creation by connect to anki
    i. First draft should just create a card with reading followed by meaning
4. CReate browser extension for interacting with japanese subtitles (This will be toughest for me as I have little experience in required techs like node.js)
5. Link browser extension to hosted backend api for querying and creating anki cards.
6. Expand above process to work with multiple accounts not just personal - more for added complexity as hosting backend will cause a cost if distributed.
7. Expand card creation to add whole sentence of collected word and potentially link to video or audio clip grab could be interesting?


The above is non exhaustive just a general aim for me to check in on as I go. This is a personal project for myself to help with both japanese learning and technical skill improvement.