# Consensus-Log-Analysis
Analyze Harmony Logs for Consensus Performance Metrics

# Usage
```console
cd log-analysis
go run . /path/to/log
```


## TODO
Make data easily digestable instead of spewing to command line.
  * Build some graphs
  * Populate some webpage with graphs
  * Or send data to jupyter/Redo project for jupyter

Use better metrics for data analysis
  * Currently just delta of simple metrics, could use many more higher order metrics as well.

Ensure this runs on real logs/further testing
  * The only logs accessible for this initial commit were local machine test logs. 
  * This was built to work with expected real-world logs, but further work is needed. Particular attention needs to be paid to when consensus fails
  * Depending on how reliable real logs are more or less conditions need to be checked in analyze.go
  * If logs are very consistent heavy optimizations can be made

Decode parsed data into structs instead of maps
  * Original commit used structs but inner, unstructured, data needed to be parsed, so maps were used instead. 
  * It is possible to override the JSON library Unmarshal function to use structs again for parsing.

Replace slice with another data structure in analyze.go
  * In the case that it's not necessary to keep each block in memory, we can use another data structure.

Improve analyzeOutput in analyze.go
  * Lots of duplicated code. Plan to change when more is known about expected runtime.