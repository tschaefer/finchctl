#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="internal/service/assets/docker-compose.yaml.tmpl"
WRITE_MODE=false

get_opts() {
    while getopts "w" opt; do
        case $opt in
            w) WRITE_MODE=true ;;
            *) echo '{"error": "Usage: '"$0"' [-w]"}' >&2; exit 2 ;;
        esac
    done
}

get_latest_tag() {
    local repo=$1
    local prefix=$2

    local url="https://registry.hub.docker.com/v2/repositories/${repo}/tags?page_size=100&ordering=last_updated"

    local latest
    latest=$(
        curl -s "$url" | \
        jq -r --arg prefix "$prefix" \
        '.results[]? | select(.name | test("^" + $prefix + "[0-9]+\\.[0-9]+\\.[0-9]+$")) | .name' | \
        sed "s/^${prefix}//" | \
        sort -V -r | \
        head -1
    )

    if [[ -n "$latest" ]]; then
        echo "${prefix}${latest}"
        return
    fi

    echo ""
}

check_image() {
    local full_image=$1

    local image_name
    local current_version
    image_name=$(echo "$full_image" | cut -d: -f1)
    current_version=$(echo "$full_image" | cut -d: -f2)

    if [[ "$image_name" == ghcr.io/* ]]; then
        echo "{
            \"image\":\"$image_name\",
            \"current\":\"$current_version\",
            \"latest\":null,
            \"status\":\"skipped\"
        }"
        return
    fi

    local prefix=""
    if [[ "$current_version" == v* ]]; then
        prefix="v"
        current_version="${current_version#v}"
    fi

    local repo
    if [[ "$image_name" != */* ]]; then
        repo="library/$image_name"
    else
        repo="$image_name"
    fi

    local latest
    latest=$(get_latest_tag "$repo" "$prefix")

    if [[ -z "$latest" ]]; then
        echo "{
            \"image\":\"$image_name\",
            \"current\":\"${prefix}${current_version}\",
            \"latest\":null,
            \"status\":\"error\"
        }"
        return
    fi

    if [[ "$current_version" == "$latest" ]]; then
        echo "{
            \"image\":\"$image_name\",
            \"current\":\"${prefix}${current_version}\",
            \"latest\":\"$latest\",
            \"status\":\"up-to-date\"
        }"
        return
    fi

    if [[ "$WRITE_MODE" == true ]]; then
        sed -i "s|${image_name}:${prefix}${current_version}|${image_name}:${latest}|g" "$COMPOSE_FILE"
        echo "{
            \"image\":\"$image_name\",
            \"current\":\"${prefix}${current_version}\",
            \"latest\":\"$latest\",
            \"status\":\"updated\"
        }"
        return
    fi

    echo "{
        \"image\":\"$image_name\",
        \"current\":\"${prefix}${current_version}\",
        \"latest\":\"$latest\",
        \"status\":\"available\"
    }"
}

main() {
    get_opts ${1+"$@"}

    local images results json_output

    readarray -t images < <(yq -r '.services[].image' "$COMPOSE_FILE")

    results=()
    for image in "${images[@]}"; do
        results+=("$(check_image "$image")")
    done

    json_output="["
    for i in "${!results[@]}"; do
        [[ $i -gt 0 ]] && json_output+=","
        json_output+="${results[$i]}"
    done
    json_output+="]"

    echo "$json_output" | jq '.'

    if echo "$json_output" | jq -e '.[] | select(.status == "available" or .status == "updated")' > /dev/null; then
        exit 0
    else
        exit 1
    fi
}

main ${1+"$@"}
