# پروژه Reflex - [نام شما]

## شماره دانشجویی
[شماره دانشجویی شما]

## توضیحات
پیاده‌سازی پروتکل Reflex در `xray-core/proxy/reflex/` انجام شده است و شامل:
- Handshake با X25519 و HKDF
- احراز هویت بر پایه UUID
- رمزنگاری frame با ChaCha20-Poly1305
- Fallback برای ترافیک غیر Reflex
- Traffic Morphing و کنترل frameهای مرتبط

## نحوه اجرا
برای اجرای تست‌های اصلی:

```bash
cd xray-core
go test ./proxy/reflex/inbound/... -timeout 180s
go test -race ./proxy/reflex/inbound/... -v
golangci-lint run ./proxy/reflex/...
```

برای بررسی پوشش:

```bash
cd xray-core
go test ./proxy/reflex/inbound -coverprofile=cover.out
go tool cover -func cover.out
```

## مشکلات و راه‌حل‌ها
- Race در تست integration با thread-safe کردن `bufferConn` برطرف شد.
- افت coverage بعد از cleanup با بازنویسی تست‌های هدفمند جبران شد.
- خطاهای lint (`errcheck`/`staticcheck`) در کد و تست‌ها رفع شدند.

