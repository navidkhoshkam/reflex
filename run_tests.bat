@echo off
REM Batch Script for running Reflex Protocol tests
REM Usage: run_tests.bat

echo ========================================
echo   Reflex Protocol Test Runner
echo ========================================
echo.

REM Change to the main project directory
cd /d E:\reflex\xray-core

echo Current directory: %CD%
echo.

REM Check if go.mod needs update
go mod verify >nul 2>&1
if errorlevel 1 (
    echo WARNING: go.mod may need update.
    echo If tests fail, try: go mod tidy
    echo.
)

REM Selection menu
echo Please choose one of the following options:
echo 1. Run all tests
echo 2. Run with Coverage
echo 3. Run with Race Detector
echo 4. Run specific test
echo 5. Run without QUIC (if go mod tidy fails)
echo 6. Exit
echo.

set /p choice="Your choice (1-6): "

if "%choice%"=="1" (
    echo.
    echo Checking go.mod...
    go mod verify >nul 2>&1
    if errorlevel 1 (
        echo.
        echo WARNING: go.mod needs update. Running go mod tidy...
        go mod tidy
        if errorlevel 1 (
            echo.
            echo ERROR: Failed to run go mod tidy. This requires internet connection.
            echo.
            echo You can try running tests without QUIC:
            echo   go test -tags "!quic" ./proxy/reflex/inbound/... -v
            pause
            exit /b 1
        )
    )
    echo.
    echo Running all tests...
    go test ./proxy/reflex/inbound/... -v
) else if "%choice%"=="2" (
    echo.
    echo Running with Coverage...
    go test -cover ./proxy/reflex/inbound/...
    echo.
    echo Generating Coverage Report...
    go test -coverprofile=coverage.out ./proxy/reflex/inbound/...
    if exist coverage.out (
        echo Coverage report saved in coverage.out
        echo To display HTML: go tool cover -html=coverage.out
    )
) else if "%choice%"=="3" (
    echo.
    echo Running with Race Detector...
    go test -race ./proxy/reflex/inbound/... -v
) else if "%choice%"=="4" (
    echo.
    echo Available tests:
    echo - TestHandshake
    echo - TestEncryptionDecryption
    echo - TestFallback
    echo - TestReplayProtection
    echo - TestTrafficProfile
    echo - TestEmptyData
    echo.
    set /p testName="Enter the test name (or pattern): "
    echo.
    echo Running test: %testName%
    go test -run %testName% ./proxy/reflex/inbound/... -v
) else if "%choice%"=="5" (
    echo.
    echo Running tests without QUIC...
    go test -tags "!quic" ./proxy/reflex/inbound/... -v
) else if "%choice%"=="6" (
    echo Exiting...
    exit
) else (
    echo Invalid option!
)

echo.
echo ========================================
echo   Tests completed
echo ========================================
pause
