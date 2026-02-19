# PowerShell Script برای اجرای تست‌های Reflex Protocol
# استفاده: .\run_tests.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Reflex Protocol Test Runner" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# تغییر به directory اصلی پروژه
$projectRoot = "E:\reflex\xray-core"
Set-Location $projectRoot

Write-Host "Current directory: $(Get-Location)" -ForegroundColor Yellow
Write-Host ""

# منوی انتخاب
Write-Host "لطفاً یکی از گزینه‌ها را انتخاب کنید:" -ForegroundColor Green
Write-Host "1. اجرای همه تست‌ها" -ForegroundColor White
Write-Host "2. اجرای با Coverage" -ForegroundColor White
Write-Host "3. اجرای با Race Detector" -ForegroundColor White
Write-Host "4. اجرای تست خاص" -ForegroundColor White
Write-Host "5. اجرای همه (Coverage + Race)" -ForegroundColor White
Write-Host "6. خروج" -ForegroundColor White
Write-Host ""

$choice = Read-Host "انتخاب شما (1-6)"

switch ($choice) {
    "1" {
        Write-Host "`nاجرای همه تست‌ها..." -ForegroundColor Cyan
        go test ./proxy/reflex/inbound/... -v
    }
    "2" {
        Write-Host "`nاجرای با Coverage..." -ForegroundColor Cyan
        go test -cover ./proxy/reflex/inbound/...
        Write-Host "`nتولید Coverage Report..." -ForegroundColor Cyan
        go test -coverprofile=coverage.out ./proxy/reflex/inbound/...
        if (Test-Path "coverage.out") {
            Write-Host "Coverage report در coverage.out ذخیره شد" -ForegroundColor Green
            Write-Host "برای نمایش HTML: go tool cover -html=coverage.out" -ForegroundColor Yellow
        }
    }
    "3" {
        Write-Host "`nاجرای با Race Detector..." -ForegroundColor Cyan
        go test -race ./proxy/reflex/inbound/... -v
    }
    "4" {
        Write-Host "`nتست‌های موجود:" -ForegroundColor Cyan
        Write-Host "- TestHandshake" -ForegroundColor White
        Write-Host "- TestEncryptionDecryption" -ForegroundColor White
        Write-Host "- TestFallback" -ForegroundColor White
        Write-Host "- TestReplayProtection" -ForegroundColor White
        Write-Host "- TestTrafficProfile" -ForegroundColor White
        Write-Host "- TestEmptyData" -ForegroundColor White
        Write-Host ""
        $testName = Read-Host "نام تست را وارد کنید (یا pattern)"
        Write-Host "`nاجرای تست: $testName" -ForegroundColor Cyan
        go test -run $testName ./proxy/reflex/inbound/... -v
    }
    "5" {
        Write-Host "`nاجرای همه تست‌ها با Coverage و Race Detector..." -ForegroundColor Cyan
        Write-Host "`n1. Coverage..." -ForegroundColor Yellow
        go test -cover ./proxy/reflex/inbound/...
        Write-Host "`n2. Race Detector..." -ForegroundColor Yellow
        go test -race ./proxy/reflex/inbound/... -v
    }
    "6" {
        Write-Host "خروج..." -ForegroundColor Yellow
        exit
    }
    default {
        Write-Host "گزینه نامعتبر!" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  تست‌ها به پایان رسید" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

