# 📓 Changelog 
## Version 0.6.1 (Aug 2026)

### bug fixes 🕷️
- fixed the laggy typing reported in [#41](https://github.com/giovannicoppola/AlfreDo/issues/41): a refresh rate of `0` made every keystroke wait for a Todoist API call. The rate is now clamped to a minimum of 1 day — use the refresh keyword (default `todoist::refresh`) for an immediate rebuild.
- no longer scatters `labelCounts`/`projectCounts` JSON files into the working directory when no data folder is configured
- logs a warning when no data folder is set (the cache can't be saved, so every run hits the API)

## Version 0.4 (Feb 2024)

### New Features 💫
- create new labels while creating tasks
- add task priority
- Added partial match checkbox in `Workflow Configuration`
   - Search projects and labels anywhere in the string. Will search from start if unchecked.
- allowing spaces in labels
- special characters `()@ # | "\& , !` are now supported in labels and project names
    except:
    1. both a space and a closed parenthesis in projects and labels
    2. `/` in projects (reserved for sections)
- possible to add time due ⏰ after date shortcuts (e.g. `2w10:13`, `2m20:45` etc.)


### bug fixes 🕷️
- would occasionally open the workflow folder after selecting a project or label
- better handling of missing projects or labels
- accepting Z timezone in due date


