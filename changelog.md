# 📓 Changelog 
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


