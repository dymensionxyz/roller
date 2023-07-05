#!/bin/bash

base_dir=$(pwd)

# define the list of directories
declare -A folders=( 
    # ["/Users/mtsitrin/Applications/dymension/rollapp"]="mtsitrin/222-evm-doesnt-work-if-no-validators-defined"
    # ["/Users/mtsitrin/Applications/dymension/settelment"]=""
    ["/Users/mtsitrin/Applications/dymension/dymension-relayer"]="" 
  )

  # define the list of directories and custom commands
declare -A commands=( 
    ["/Users/mtsitrin/Applications/dymension/rollapp"]="make build_evm"
)

# define the list of operating systems
oslist=("darwin" "linux")

# define the list of architectures
archlist=("amd64" "arm64")

# loop through the directories
for folder in "${!folders[@]}"; do
    branch=${folders[$folder]}
    echo "Entering folder $folder"
    
    cd $folder
    rm ./build/*
    
 # checkout the specific branch or tag if one is defined
    if [[ -n "$branch" ]]; then
        echo "Checking out $branch"
        git checkout $branch
    fi

    # loop through the os and arch
    for os in "${oslist[@]}"; do
        for arch in "${archlist[@]}"; do
            echo "Building for $os $arch"

            # if there is a custom command for this directory, run it
            if [[ -n "${commands[$folder]}" ]]; then
                env GOOS=$os GOARCH=$arch ${commands[$folder]}
            else
                env GOOS=$os GOARCH=$arch make build
            fi

            read -p "Press any key to resume ..."

            # move the build outputs to a specific folder
            mkdir -p $base_dir/outputs/$os-$arch
            mv ./build/* $base_dir/outputs/$os-$arch/
            rm ./build/*
        done
    done

    echo "Exiting folder $folder"
done
cd $base_dir
