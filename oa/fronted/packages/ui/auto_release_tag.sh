#!/bin/bash
echo "Please enter the module name (ui):"
read name

# 如果用户输入为空，则默认设置为 "api"
if [ -z "$name" ]; then
  name="ui"
  # cp -f ./configs/app_online.json ./configs/app.json
fi

# Define an array of module names
modules=("ui")

currentTime=$(date "+%Y%m%d%H%M")


if [[ "$name" == "all" ]]; then
    # Execute for all modules
    for module in "${modules[@]}"; do
        git tag "release-$currentTime-$module"
    done
else
    # Execute for the specified module
    git tag "release-$currentTime-$name"
fi

# Push the tags to the remote repository
git push origin --tags
