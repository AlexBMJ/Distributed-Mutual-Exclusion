# Distributed-Mutual-Exclusion
Mandatory Exercise 2 - Distributed Systems

## Running the application
```bash
# Run the bash script
./run.sh
```

All docker containers map to the log.txt file, where they output their name and the current token value. The token increases by one when it is passed to another node. Since we solved the mutual exclusion problem, there is no overlapping when writing to the same file. The run script uses tail to follow the continuously updated log.