# پروژه Reflex

این پروژه توسط **سه نفر** به صورت گروهی پیاده‌سازی شده است.

---

## نوید خوش کام

### شماره دانشجویی
400108825

### توضیحات
پیاده‌سازی کدهای خواسته شده در:
- **[Step 1 - ساختار اولیه](docs/step1-basic.md):** ساختار پکیج reflex، تعریف config.proto و تولید pb.go، handler اولیه inbound و outbound
- **[Step 2 - Handshake](docs/step2-handshake.md):** تبادل کلید X25519، استخراج کلید جلسه با HKDF، احراز هویت با UUID، مدیریت خطا

### نحوه اجرا
```bash
cd xray-core
go build -o xray ./main
go test ./proxy/reflex/inbound/... -timeout 180s
go test -race ./proxy/reflex/inbound/... -v
```

برای استفاده با config:
```bash
xray run -c ../config.example.json
```

### مشکلات و راه‌حل‌ها
- ثبت پروتکل در Xray و سازگاری با infra/conf برای parse کردن config JSON
- تطبیق نوع بازگشتی `New` با `proxy.Inbound` (به‌جای `proxy.InboundHandler`)

---

## محمدرضا ضیاء

### شماره دانشجویی
400108871

### توضیحات
پیاده‌سازی کدهای خواسته شده در:
- **[Step 4 - Fallback و Multiplexing](docs/step4-fallback.md):** تشخیص پروتکل با bufio.Peek، Fallback به وب‌سرور، Multiplexing روی یک پورت، wrapper پیش‌بارگذاری (preloadedConn) برای حفظ بایت‌های peek شده

### نحوه اجرا
```bash
cd xray-core
go build -o xray ./main
go test ./proxy/reflex/inbound/... -timeout 180s
go test -race ./proxy/reflex/inbound/... -v
```

برای استفاده با config:
```bash
xray run -c ../config.example.json
```

### مشکلات و راه‌حل‌ها
- مصرف نشدن بایت‌های peek شده هنگام fallback: استفاده از `preloadedConn` که reader را wrap می‌کند تا بایت‌های peek شده به وب‌سرور ارسال شوند
- هماهنگی timeout و policy از core برای `buf.Copy` در fallback

---

## تبسم فتحی

### شماره دانشجویی
400108893

### توضیحات
پیاده‌سازی کدهای خواسته شده در:
- **[Step 3 - رمزنگاری و پردازش بسته‌ها](docs/step3-encryption.md):** ساختار Frame، رمزنگاری با ChaCha20-Poly1305، خواندن/نوشتن frame، محافظت در برابر replay با nonceهای ترتیبی

### نحوه اجرا
```bash
cd xray-core
go build -o xray ./main
go test ./proxy/reflex/inbound/... -timeout 180s
go test -race ./proxy/reflex/inbound/... -v
```

برای استفاده با config:
```bash
xray run -c ../config.example.json
```

### مشکلات و راه‌حل‌ها
- تطبیق اندازه nonce (12 بایت) با انتظار ChaCha20-Poly1305
- یکپارچه‌سازی frame با Xray dispatcher و bidirectional forwarding

---

## کار گروهی (Step 5 و سایر)

کدهای **[Step 5 - قابلیت‌های پیشرفته](docs/step5-advanced.md)** (Traffic Morphing، TrafficProfile، PADDING_CTRL و TIMING_CTRL، پروفایل‌های آماده مثل youtube/zoom) و سایر بخش‌های مشترک را هر سه نفر با هم پیاده‌سازی کردیم.

---

## نحوه اجرا (مشترک)

```bash
cd xray-core
go build -o xray ./main
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

مثال پیکربندی در `config.example.json` موجود است.

---

## مشکلات و راه‌حل‌ها (مشترک)

- **Race در تست integration:** با thread-safe کردن `bufferConn` (استفاده از `sync.Mutex`) برطرف شد.
- **افت coverage بعد از cleanup:** با بازنویسی تست‌های هدفمند و تفکیک فایل‌های تست جبران شد.
- **خطاهای lint:** `errcheck` و `staticcheck` در کد و تست‌ها رفع شدند.
