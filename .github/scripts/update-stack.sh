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
    local name=$1
    local repo=$2
    local current_version=$3
    local prefix=$4
    local image_name=$5

    local latest
    latest=$(get_latest_tag "$repo" "$prefix")

    if [[ -z "$latest" ]]; then
        echo "{
            \"name\":\"$name\",
            \"current\":\"$current_version\",
            \"latest\":null,
            \"status\":\"error\"
        }"
        return
    fi

    if [[ "$current_version" == "$latest" ]]; then
        echo "{
            \"name\":\"$name\",
            \"current\":\"$current_version\",
            \"latest\":\"$latest\",
            \"status\":\"up-to-date\"
        }"
        return
    fi

    if [[ "$WRITE_MODE" == true ]]; then
        sed -i "s|${image_name}:${current_version}|${image_name}:${latest}|g" "$COMPOSE_FILE"
        echo "{
            \"name\":\"$name\",
            \"current\":\"$current_version\",
            \"latest\":\"$latest\",
            \"status\":\"updated\"
        }"
        return
    fi

    echo "{
        \"name\":\"$name\",
        \"current\":\"$current_version\",
        \"latest\":\"$latest\",
        \"status\":\"available\"
    }"
}

main() {
    get_opts ${1+"$@"}

    local grafana_version loki_version traefik_version alloy_version mimir_version pyroscope_version
    local results json_output

    grafana_version=$(grep -oP 'grafana/grafana:\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")
    loki_version=$(grep -oP 'grafana/loki:\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")
    traefik_version=$(grep -oP 'traefik:v\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")
    alloy_version=$(grep -oP 'grafana/alloy:v\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")
    mimir_version=$(grep -oP 'grafana/mimir:\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")
    pyroscope_version=$(grep -oP 'grafana/pyroscope:\K[0-9]+\.[0-9]+\.[0-9]+' "$COMPOSE_FILE")

    results=()
    results+=("$(check_image "grafana" "grafana/grafana" "$grafana_version" "" "grafana/grafana")")
    results+=("$(check_image "loki" "grafana/loki" "$loki_version" "" "grafana/loki")")
    results+=("$(check_image "traefik" "library/traefik" "v$traefik_version" "v" "traefik")")
    results+=("$(check_image "alloy" "grafana/alloy" "v$alloy_version" "v" "grafana/alloy")")
    results+=("$(check_image "mimir" "grafana/mimir" "$mimir_version" "" "grafana/mimir")")
    results+=("$(check_image "pyroscope" "grafana/pyroscope" "$pyroscope_version" "" "grafana/pyroscope")")

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
