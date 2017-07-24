# This file goes into /etc/profile.d - it will enable snap completion 
# 

# Check for interactive bash and that we haven't already been sourced.
if [ -n "$BASH_VERSION" -a -n "$PS1" -a -z "$SNAP_COMPLETE" ]; then
    # check if completion is on
    if shopt -q progcomp && [ -r /usr/lib/snapd/complete.sh ]; then
        # Source completion code.
        . /usr/lib/snapd/complete.sh
    fi

fi
