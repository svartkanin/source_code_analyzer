# Code analyzer

A simple code analyzer tool for code repositories. The output shows an overall analysis and also per file in a JSON format.

## Dependencies
The code has the following dependencies

```github.com/gookit/config/json```

which can be installed with 

```
dep ensure
```

## Build and run
Build the executable
```go build -o analyzer```

Modify the `config.json` file and set a source code directory, then run the analyzer with

```./analyzer --config config.json --output results.json ```

## Example output

```
{
  "java": {
    "TotalFiles": 71,
    "TotalResults": {
      "Lines": 5108,
      "CodeLines": 4221,
      "Comments": 825,
      "Keywords": 3464
    },
    "Files": {
      "/tmp/metasploit-framework/external/source/exploits/CVE-2008-5353/src/msf/x/AppletX.java": {
        "Lines": 76,
        "CodeLines": 65,
        "Comments": 22,
        "Keywords": 33
      },
      "/tmp/metasploit-framework/external/source/exploits/CVE-2008-5353/src/msf/x/LoaderX.java": {
        "Lines": 96,
        "CodeLines": 77,
        "Comments": 19,
        "Keywords": 64
      },
      ...
    }
  },
  "python": {
    "TotalFiles": 28,
    "TotalResults": {
      "Lines": 4476,
      "CodeLines": 3881,
      "Comments": 378,
      "Keywords": 2363
    },
    "Files": {
      "/tmp/metasploit-framework/data/exploits/CVE-2015-1130/exploit.py": {
        "Lines": 73,
        "CodeLines": 56,
        "Comments": 10,
        "Keywords": 29
      },
      "/tmp/metasploit-framework/external/source/exploits/splunk/upload_app_exec/bin/msf_exec.py": {
        "Lines": 15,
        "CodeLines": 11,
        "Comments": 0,
        "Keywords": 6
      },
      "/tmp/metasploit-framework/external/source/shellcode/windows/x64/build.py": {
        "Lines": 106,
        "CodeLines": 106,
        "Comments": 23,
        "Keywords": 72
      },
      ...
    }
  }
}
```