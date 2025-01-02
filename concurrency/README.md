# Concurrency Test

## Description
This is a simple test to compare the performance of sequential, concurrent, and fair concurrent execution of a simple task. Counting no of prime numbers in a given range.

## Experiment: Sequential vs Concurrent vs Fair Concurrent

### Results

#### Range 0 - 100 Millions

---
Time taken:  31.491680292s
Total prime numbers:  5761455
---
Thread Id 0, Time taken: 3.264567583s
Thread Id 1, Time taken: 5.010849458s
Thread Id 2, Time taken: 5.873900375s
Thread Id 3, Time taken: 6.766418209s
Thread Id 4, Time taken: 7.385691625s
Thread Id 5, Time taken: 7.643053833s
Thread Id 6, Time taken: 7.884509417s
Thread Id 7, Time taken: 8.343699166s
Time taken:  8.343806s
---
Thread Id 7, Time taken: 8.077165708s
Thread Id 1, Time taken: 8.077162042s
Thread Id 0, Time taken: 8.077062958s
Thread Id 6, Time taken: 8.077023875s
Thread Id 4, Time taken: 8.077010542s
Thread Id 2, Time taken: 8.077120458s
Thread Id 3, Time taken: 8.077143875s
Thread Id 5, Time taken: 8.077068167s
Time taken:  8.077243s

#### Range 0 - 1 Billion
---
Thread Id 0, Time taken: 1m31.227307166s
Thread Id 1, Time taken: 2m25.586779875s
Thread Id 2, Time taken: 2m55.007260334s
Thread Id 3, Time taken: 3m13.885404958s
Thread Id 4, Time taken: 3m27.558678125s
Thread Id 5, Time taken: 3m38.896917542s
Thread Id 6, Time taken: 3m49.248288709s
Thread Id 7, Time taken: 3m58.624386625s
Time taken:  3m58.624518167s
---
Thread Id 3, Time taken: 3m32.381040791s
Thread Id 6, Time taken: 3m32.381066625s
Thread Id 2, Time taken: 3m32.3813145s
Thread Id 7, Time taken: 3m32.381174917s
Thread Id 1, Time taken: 3m32.381108291s
Thread Id 0, Time taken: 3m32.381114875s
Thread Id 4, Time taken: 3m32.381179042s
Thread Id 5, Time taken: 3m32.381179333s
Time taken:  3m32.38142175s

