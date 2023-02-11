#!/bin/bash


for script in *-*.sh; do
    if ! [ -x ./"$script" ]; then
        echo >&2 "$script is not executable, ignoring"
        continue
    fi
    ./"$script"
    exitcode="$?"
    case "$exitcode" in
        0) echo >&2 "$script configured the screen"; break ;;
        2) echo >&2 "$script was not applicable to the current setup" ;;
        *) echo >&2 "$script failed with exit code $exitcode, ignoring" ;;
    esac
done
