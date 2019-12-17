window.BENCHMARK_DATA = {
  "lastUpdate": 1576613364374,
  "repoUrl": "https://github.com/Brian-Williams/retry",
  "entries": {
    "Go Benchmark": [
      {
        "commit": {
          "author": {
            "email": "Brian-Williams@users.noreply.github.com",
            "name": "Brian",
            "username": "Brian-Williams"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6d745c668572dac0db13c3e9cc957db50df2fcd9",
          "message": "Update go.yml\n\nUse github token for now.",
          "timestamp": "2019-12-17T15:08:41-05:00",
          "tree_id": "15ab65775c1a5586c4d986172eb9a897a1052130",
          "url": "https://github.com/Brian-Williams/retry/commit/6d745c668572dac0db13c3e9cc957db50df2fcd9"
        },
        "date": 1576613363926,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkDo10",
            "value": 260,
            "unit": "ns/op",
            "extra": "4756117 times\n2 procs"
          },
          {
            "name": "BenchmarkDoWithConfigurer10",
            "value": 240,
            "unit": "ns/op",
            "extra": "5029131 times\n2 procs"
          },
          {
            "name": "BenchmarkDo100",
            "value": 2120,
            "unit": "ns/op",
            "extra": "568138 times\n2 procs"
          },
          {
            "name": "BenchmarkDoWithConfigurer100",
            "value": 2022,
            "unit": "ns/op",
            "extra": "601276 times\n2 procs"
          },
          {
            "name": "BenchmarkDo10000",
            "value": 210818,
            "unit": "ns/op",
            "extra": "5004 times\n2 procs"
          },
          {
            "name": "BenchmarkDoWithConfigurer10000",
            "value": 193987,
            "unit": "ns/op",
            "extra": "6332 times\n2 procs"
          }
        ]
      }
    ]
  }
}