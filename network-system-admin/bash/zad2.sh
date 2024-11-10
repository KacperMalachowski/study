#!/bin/bash

# Function to fetch correct city name and weather URL
function searchCity() {
  city=$1

  cities=$(curl -s "https://pogoda.onet.pl/ajax/search?type=&query=$city")
}


searchCity "warszawa"
cityName=$(echo $cities | jq -r '.suggestions[0].value')
weatherURL=$(echo $cities | jq -r '.suggestions[0].data.url')
weather=$(curl -s "https://pogoda.onet.pl$weatherURL")
temp=$(echo $weather | grep -oP '<div class="temp">\K[0-9]+(?=<span class="deg">)')
iconURL=$(echo $weather | grep -oP '<span class="iconHolder">\s*<img[^>]*src="\K[^"]+' | head -n 1)
iconAlt=$(echo $weather | grep -oP '<span class="iconHolder">\s*<img[^>]*alt="\K[^"]+' | head -n 1)
rain=$(echo $weather | grep -oP '<span class="restParamValue">\K[^"]+(?=</span>)' | head -n 1)
wind=$(echo $weather | grep -oP '<span class="restParamValue">\K[^"]+(?=<span class="windDirectionArrow")' | head -n 1)
qnh=$(echo $weather | grep -oP '<span class="restParamLabel">Ciśnienie atmosferyczne</span>\s*<span class="restParamValue">\K[^"]+(?=</span>)' | head -n 1)
humidity=$(echo $weather | grep -oP '<span class="restParamLabel">Wilgotność</span>\s*<span class="restParamValue">\K[^"]+(?=</span>)' | head -n 1)

curl -s -o 1.png "$iconURL"
echo $cityName
echo "$temp°C"
echo $iconAlt
echo "Deszcz: $rain"
echo "Wiatr: $wind"
echo "Ciśnienie: $qnh"
echo "Wilgotność: $humidity"
