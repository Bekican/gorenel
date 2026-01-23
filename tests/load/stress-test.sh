#!/bin/bash
# Progressive Stress Test for Gorenel
# Gradually increases load until failure point is found

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

URL="${1:-https://test.tunnel-project.xyz}"
START_CONCURRENCY="${2:-10}"
MAX_CONCURRENCY="${3:-1000}"
STEP="${4:-50}"
DURATION="${5:-30}"

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}Gorenel Progressive Stress Test${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo "Target: ${URL}"
echo "Starting concurrency: ${START_CONCURRENCY}"
echo "Max concurrency: ${MAX_CONCURRENCY}"
echo "Step: ${STEP}"
echo "Duration per step: ${DURATION}s"
echo ""

mkdir -p results
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="results/stress_${TIMESTAMP}.csv"

echo "Concurrency,RPS,AvgLatency,P95Latency,ErrorRate" > ${RESULT_FILE}

current=${START_CONCURRENCY}
while [ ${current} -le ${MAX_CONCURRENCY} ]; do
    echo -e "${YELLOW}Testing with ${current} concurrent connections...${NC}"
    
    # Run k6 for this load level
    k6 run --duration ${DURATION}s \
           --vus ${current} \
           --out json=results/k6_${current}.json \
           --quiet \
           tests/load/k6-test.js \
           2>&1 | tee /tmp/k6_output.txt
    
    # Extract metrics
    RPS=$(grep "http_reqs" /tmp/k6_output.txt | awk '{print $2}')
    AVG_LATENCY=$(grep "http_req_duration.*avg" /tmp/k6_output.txt | awk '{print $3}')
    P95_LATENCY=$(grep "http_req_duration.*p(95)" /tmp/k6_output.txt | awk '{print $3}')
    ERROR_RATE=$(grep "errors" /tmp/k6_output.txt | awk '{print $2}')
    
    echo "${current},${RPS},${AVG_LATENCY},${P95_LATENCY},${ERROR_RATE}" >> ${RESULT_FILE}
    
    echo -e "  RPS: ${RPS}"
    echo -e "  Avg Latency: ${AVG_LATENCY}ms"
    echo -e "  P95 Latency: ${P95_LATENCY}ms"
    echo -e "  Error Rate: ${ERROR_RATE}%"
    echo ""
    
    # Check if system is degrading
    if (( $(echo "${ERROR_RATE} > 5" | bc -l) )); then
        echo -e "${RED}Error rate exceeded 5% at ${current} concurrent connections${NC}"
        echo -e "${RED}Breaking point found!${NC}"
        break
    fi
    
    if (( $(echo "${P95_LATENCY} > 2000" | bc -l) )); then
        echo -e "${YELLOW}P95 latency exceeded 2000ms at ${current} concurrent connections${NC}"
        echo -e "${YELLOW}Performance degradation detected${NC}"
    fi
    
    current=$((current + STEP))
    sleep 5  # Cool down between steps
done

echo ""
echo -e "${GREEN}Stress test completed!${NC}"
echo -e "${GREEN}Results saved to: ${RESULT_FILE}${NC}"

# Generate summary
echo ""
echo -e "${GREEN}Performance Summary:${NC}"
tail -1 ${RESULT_FILE}

# Generate plot if gnuplot available
if command -v gnuplot &> /dev/null; then
    cat > /tmp/gnuplot_stress.txt << 'EOF'
set terminal png size 1400,800
set output "results/stress_plot_${TIMESTAMP}.png"
set title "Gorenel Stress Test - Performance Under Load"
set xlabel "Concurrent Connections"
set ylabel "Requests per Second"
set y2label "Latency (ms)"
set grid
set datafile separator ","
set ytics nomirror
set y2tics
plot "results/stress_${TIMESTAMP}.csv" using 1:2 with linespoints title "RPS" axis x1y1, \
     "" using 1:4 with linespoints title "P95 Latency" axis x1y2
EOF
    
    gnuplot /tmp/gnuplot_stress.txt
    echo -e "${GREEN}Performance graph: results/stress_plot_${TIMESTAMP}.png${NC}"
fi