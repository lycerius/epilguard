# Epilguard
Epilguard analyzes for photosensitive content in videos, specifically flashing content. The project is compliant with [ITU-R BT.1702](https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.1702-0-200502-I!!PDF-E.pdf), which details how to identify photosensitive hazards in video content.
## Using Epilguard
You can download the release binaries of epilguard for your platform [here](https://github.com/lycerius/epilguard/releases/latest). Epilguard depends on FFMpeg to decode videos so you will need to download a build of [FFMpeg](https://www.ffmpeg.org/) for your platform and add it to your `$PATH`.


To use epilguard:
``` sh
#!/bin/bash
$ epilguard
epilguard [options] video

  -buffer-size uint
        Sets the size of the lookahead framebuffer, must be > 0 (default 30)
  -report-dir string
        directory to write report files to (default $cwd)
```

[input-file] is the video to analyze, [csv-export-directory] is where you would like to write the report artifacts to.
## Creating Reports
To create a hazard report for a video:
``` sh
$ epilguard -report-dir=report/directory videoname
```

This will analyze video.mp4 and create the hazard report in report/directory. The report consists of 4 separate files:

``` tree
report/directory/
├── [timestamp]-[videoname]-Accumulation.csv
├── [timestamp]-[videoname]-Flashes.csv
├── [timestamp]-[videoname]-FrameFlashes.csv
└── [timestamp]-[videoname]-Report.json
```

_Description of Artifacts_
* **Accumulation** - How the brightness accumulated over frames before inversion.
* **Flashes** - A compressed version of accumulation that details the maximum brightness achieved over X frames before inversion
* **FrameFlashes** - Like flashes, but uses frame indexes instead of frame count
* **Report** - The descriptive hazard report which describes each hazard and where it started/ended in the video.

*Note: You can plot the CSV files using a common plotting utility (such as Excel or PyPlot) to visualize the hazard breakdown.*
## Hazard Report
A Hazard Report is a JSON file that contains a number of hazards and descriptions that explain why they were considered hazardous.

Example Hazard Report:
```
{
  
    "createdOn": DateString,
    "hazards": [
        {
            "start": number,
            "end": number,
            "hazardType": string
        },
        ...
    ]
}
```



## Building Epilguard from Source
### Installing Golang
Install the latest build of Golang for your platform.
1. Download Golang [here](https://golang.org/dl/)
2. Follow this [guide](https://golang.org/doc/install)
### Cloning and Building Epilguard
``` sh
$ go get -u github.com/lycerius/epilguard
```

If cloning was successful, you can now execute Epilguard as normal
``` sh
$ epilguard
epilguard [options] video

  -buffer-size uint
        Sets the size of the lookahead framebuffer, must be > 0 (default 30)
  -report-dir string
        directory to write report files to (default $cwd)
```
