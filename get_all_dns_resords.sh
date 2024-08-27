#!/bin/bash

# Load credentials from external file
if [ -f "./cloudflare_credentials.sh" ]; then
    source ./cloudflare_credentials.sh
else
    echo "Error: cloudflare_credentials.sh file not found!"
    echo "Please create this file with the following content:"
    echo "AUTH_EMAIL='your_email@example.com'"
    echo "AUTH_KEY='your_api_key'"
    echo "ZONE_ID='your_zone_id'"
    exit 1
fi

# Function to add DNS record
add_dns_record() {
    local type=$1
    local name=$2
    local content=$3
    local ttl=$4

    curl -X POST "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
         -H "X-Auth-Email: ${AUTH_EMAIL}" \
         -H "X-Auth-Key: ${AUTH_KEY}" \
         -H "Content-Type: application/json" \
         --data '{
           "type":"'"${type}"'",
           "name":"'"${name}"'",
           "content":"'"${content}"'",
           "ttl":'"${ttl}"'
         }'
}

# Function to get all DNS records
get_all_dns_records() {
    curl -X GET "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records" \
         -H "X-Auth-Email: ${AUTH_EMAIL}" \
         -H "X-Auth-Key: ${AUTH_KEY}" \
         -H "Content-Type: application/json"
}

# Function to get DNS records of a specific type
get_dns_records_by_type() {
    local type=$1
    curl -X GET "https://api.cloudflare.com/client/v4/zones/${ZONE_ID}/dns_records?type=${type}&per_page=1000" \
         -H "X-Auth-Email: ${AUTH_EMAIL}" \
         -H "X-Auth-Key: ${AUTH_KEY}" \
         -H "Content-Type: application/json"
}

# Function to process and display DNS records
process_dns_records() {
    local json_data=$1
    local type=$2
    echo "DNS Records${type:+ of type $type}:"
    echo "------------"
    echo "$json_data" | jq -r '.result[] | "\(.type) | \(.name) | \(.content) | TTL: \(.ttl) | Proxied: \(.proxied)"'
    
    echo -e "\nSummary:"
    echo "--------"
    echo "$json_data" | jq -r '.result[] | .type' | sort | uniq -c | awk '{print $2 ": " $1}'
    
    total_records=$(echo "$json_data" | jq '.result | length')
    echo "Total records: $total_records"
}

# Main execution
case "$1" in
  "add")
    add_dns_record "$2" "$3" "$4" "$5"
    ;;
  "get_all")
    records_json=$(get_all_dns_records)
    process_dns_records "$records_json"
    ;;
  "get")
    if [ -z "$2" ]; then
      echo "Error: DNS record type is required for 'get' command."
      echo "Usage: $0 get <type>"
      exit 1
    fi
    records_json=$(get_dns_records_by_type "$2")
    process_dns_records "$records_json" "$2"
    ;;
  *)
    echo "Usage: $0 {add|get_all|get <type>}"
    echo "Examples:"
    echo "  $0 add A example.com 192.0.2.1 3600"
    echo "  $0 get_all"
    echo "  $0 get A"
    exit 1
    ;;
esac