#!/bin/bash
# Apache Bench Load Test for Gorenel
# Usage: ./ab-test.sh [URL] [concurrency] [requests]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
URL="${1:-https://test.tunnel-project.xyz}"
CONCURRENCY="${2:-100}"
REQUESTS="${3:-10000}"
TIMEOUT=30

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}Gorenel Load Test - Apache Bench${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo -e "Target URL:   ${YELLOW}${URL}${NC}"
echo -e "Concurrency:  ${YELLOW}${CONCURRENCY}${NC}"
echo -e "Requests:     ${YELLOW}${REQUESTS}${NC}"
echo -e "Timeout:      ${YELLOW}${TIMEOUT}s${NC}"
echo ""

# Check if ab is installed
if ! command -v ab &> /dev/null; then
    echo -e "${RED}Error: Apache Bench (ab) not found${NC}"
    echo "Install: sudo apt-get install apache2-utils"
    exit 1
fi

# Create results directory
mkdir -p results
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="results/ab_test_${TIMESTAMP}.txt"

echo -e "${YELLOW}Starting load test...${NC}"
echo ""

# Run Apache Bench
ab -n ${REQUESTS} \
   -c ${CONCURRENCY} \
   -s ${TIMEOUT} \
   -g results/gnuplot_${TIMESTAMP}.tsv \
   -e results/csv_${TIMESTAMP}.csv \
   ${URL}/ | tee ${RESULT_FILE}

echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}Test Summary${NC}"
echo -e "${GREEN}======================================${NC}"

# Extract key metrics
TOTAL_TIME=$(grep "Time taken for tests:" ${RESULT_FILE} | awk '{print $5}')
RPS=$(grep "Requests per second:" ${RESULT_FILE} | awk '{print $4}')
MEAN_TIME=$(grep "Time per request:" ${RESULT_FILE} | head -1 | awk '{print $4}')
FAILED=$(grep "Failed requests:" ${RESULT_FILE} | awk '{print $3}')

echo -e "Total Time:        ${YELLOW}${TOTAL_TIME}s${NC}"
echo -e "Requests/second:   ${YELLOW}${RPS}${NC}"
echo -e "Mean Time/Request: ${YELLOW}${MEAN_TIME}ms${NC}"
echo -e "Failed Requests:   ${YELLOW}${FAILED}${NC}"

# Calculate success rate
SUCCESS_RATE=$(echo "scale=2; (($REQUESTS - $FAILED) / $REQUESTS) * 100" | bc)
echo -e "Success Rate:      ${YELLOW}${SUCCESS_RATE}%${NC}"

echo ""
echo -e "${GREEN}Results saved to: ${RESULT_FILE}${NC}"
echo -e "${GREEN}CSV data: results/csv_${TIMESTAMP}.csv${NC}"
echo -e "${GREEN}Gnuplot data: results/gnuplot_${TIMESTAMP}.tsv${NC}"

# Generate gnuplot graph if gnuplot is available
if command -v gnuplot &> /dev/null; then
    echo ""
    echo -e "${YELLOW}Generating response time graph...${NC}"
    
    cat > /tmp/gnuplot_cmd.txt << EOF
set terminal png size 1200,800
set output "results/graph_${TIMESTAMP}.png"
set title "Gorenel Load Test - Response Time Distribution"
set xlabel "Request Number"
set ylabel "Response Time (ms)"
set grid
set datafile separator "\t"
plot "results/gnuplot_${TIMESTAMP}.tsv" using 10 with lines title "Response Time"
EOF
    
    gnuplot /tmp/gnuplot_cmd.txt
    echo -e "${GREEN}Graph saved to: results/graph_${TIMESTAMP}.png${NC}"
fi

echo ""
echo -e "${GREEN}Test completed successfully!${NC}"

# Check thresholds
if (( $(echo "$SUCCESS_RATE < 95" | bc -l) )); then
    echo -e "${RED}WARNING: Success rate below 95%${NC}"
    exit 1
fi

if (( $(echo "$MEAN_TIME > 500" | bc -l) )); then
    echo -e "${YELLOW}WARNING: Mean response time above 500ms${NC}"
fi

exit 0