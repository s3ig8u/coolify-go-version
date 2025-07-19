#!/bin/bash

# Test runner script for Coolify Go
# This script runs all tests and provides a summary

set -e

echo "🧪 Running Coolify Go Tests"
echo "============================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to run tests for a specific package
run_package_tests() {
    local package_name=$1
    local test_path=$2
    
    echo -e "\n${YELLOW}Testing ${package_name}...${NC}"
    
    if [ -d "$test_path" ]; then
        cd "$test_path"
        
        # Run tests and capture output
        test_output=$(go test -v ./... 2>&1)
        test_exit_code=$?
        
        # Parse test results
        if [ $test_exit_code -eq 0 ]; then
            echo -e "${GREEN}✅ ${package_name} tests passed${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}❌ ${package_name} tests failed${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        # Count tests
        test_count=$(echo "$test_output" | grep -c "=== RUN" || echo "0")
        TOTAL_TESTS=$((TOTAL_TESTS + test_count))
        
        # Show test output
        echo "$test_output"
        
        cd - > /dev/null
    else
        echo -e "${YELLOW}⚠️  No tests found for ${package_name}${NC}"
        SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
    fi
}

# Function to run integration tests
run_integration_tests() {
    echo -e "\n${YELLOW}Running Integration Tests...${NC}"
    
    # Check if Docker is available
    if command -v docker &> /dev/null; then
        echo "🐳 Docker is available - running integration tests"
        
        # Run Docker integration tests
        run_package_tests "Docker Integration" "internal/docker"
        
        # Run Deployment Engine tests
        run_package_tests "Deployment Engine" "internal/deployment"
        
    else
        echo -e "${YELLOW}⚠️  Docker not available - skipping integration tests${NC}"
        SKIPPED_TESTS=$((SKIPPED_TESTS + 2))
    fi
}

# Function to run unit tests
run_unit_tests() {
    echo -e "\n${YELLOW}Running Unit Tests...${NC}"
    
    # Run unit tests for all packages
    for dir in internal/*/; do
        if [ -d "$dir" ]; then
            package_name=$(basename "$dir")
            run_package_tests "$package_name" "$dir"
        fi
    done
}

# Function to run all tests
run_all_tests() {
    echo -e "\n${YELLOW}Running All Tests...${NC}"
    
    # Run tests for the entire project
    test_output=$(go test -v ./... 2>&1)
    test_exit_code=$?
    
    if [ $test_exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ All tests passed${NC}"
    else
        echo -e "${RED}❌ Some tests failed${NC}"
    fi
    
    echo "$test_output"
}

# Function to show test summary
show_summary() {
    echo -e "\n${YELLOW}Test Summary${NC}"
    echo "============"
    echo -e "Total Tests: ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
    echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"
    echo -e "${YELLOW}Skipped: ${SKIPPED_TESTS}${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests completed successfully!${NC}"
        exit 0
    else
        echo -e "\n${RED}💥 Some tests failed. Please check the output above.${NC}"
        exit 1
    fi
}

# Main execution
main() {
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}Error: go.mod not found. Please run this script from the project root.${NC}"
        exit 1
    fi
    
    # Parse command line arguments
    case "${1:-all}" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "docker")
            run_package_tests "Docker Integration" "internal/docker"
            ;;
        "deployment")
            run_package_tests "Deployment Engine" "internal/deployment"
            ;;
        "all"|*)
            run_all_tests
            ;;
    esac
    
    show_summary
}

# Run main function
main "$@" 