#!/bin/bash

function entrypoint() {
    case $1 in
        bootstrap)
            return 0
            ;;
        env)
            env | sort
            return 0
            ;;
        web)
            ;;
        *)
            echo "usage: entrypoint.sh <bootstrap|web|env>"
            return 0
            ;;
    esac

    exec /srv/azure_arm_proxy/azure_v2/azure_v2 --listen=":8083" --prefix="/azure_v2" > /dev/null
    #exec /srv/azure_arm_proxy/azure_v2/azure_v2 --listen="localhost:8084" --prefix="/azure_v2" > /dev/null
    #exec /srv/azure_arm_proxy/azure_v2/azure_v2 --listen="localhost:8085" --prefix="/azure_v2" > /dev/null
}

entrypoint $@