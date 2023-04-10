#!/usr/bin/env python3

import sys


# AlfreDo – a Todoist workflow
# Partly cloudy ⛅️  🌡️+31°F (feels +28°F, 82%) 🌬️↘4mph 🌗 2022-01-25 Tue 9:05AM

def log(s, *args):
    if args:
        s = s % args
    print(s, file=sys.stderr)




