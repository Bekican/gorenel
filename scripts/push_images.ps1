# gorenel-build-push.ps1
# This script builds and pushes Docker images to a registry.

# --- CONFIGURATION ---
$DOCKER_USER = "bekican" # Kullanici adini bekican olarak ayarladim
$PROJECT_NAME = "gorenel"
$TAG = "latest"

# --- BUILD & PUSH FUNCTIONS ---
function Build-And-Push($serviceName, $contextPath, $dockerfilePath) {
    $imageName = "${DOCKER_USER}/${PROJECT_NAME}-${serviceName}:${TAG}"
    
    # Varsayilan degeri degistirip degistirmedigini kontrol eder
    if ($DOCKER_USER -eq "your_dockerhub_username") {
        Write-Host "HATA: Lütfen script icindeki `$DOCKER_USER degiskenini kendi Docker Hub adinizla güncelleyin!" -ForegroundColor Red
        exit
    }

    Write-Host "`n>>> Building $imageName..." -ForegroundColor Green
    docker build -t $imageName -f $dockerfilePath $contextPath
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "`n[!] HATA: $serviceName build edilemedi. Islem durduruldu." -ForegroundColor Red
        exit
    }

    Write-Host ">>> Pushing $imageName..." -ForegroundColor Green
    docker push $imageName
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "`n[!] HATA: $imageName Docker Hub'a yüklenemedi. Giris yaptiginizdan ve kullanici adinizin dogru oldugundan emin olun." -ForegroundColor Red
        exit
    }
}

# --- EXECUTION ---
Write-Host "Starting build and push process for $PROJECT_NAME..." -ForegroundColor Cyan

# 1. Server (Go Backend)
Build-And-Push "server" "." "Dockerfile"

# 2. Dashboard (Next.js Frontend)
Build-And-Push "dashboard" "./web-dashboard" "./web-dashboard/Dockerfile"

# 3. ML Engine
Build-And-Push "ml-engine" "./services/ml" "./services/ml/Dockerfile"

# 4. Caddy (Custom Build)
Build-And-Push "caddy" "." "Dockerfile.caddy"

Write-Host "`n[TEBRIKLER] Tüm imajlar basariyla yüklendi!" -ForegroundColor Green
Write-Host "Simdi OCI sunucuna gidip su komutlari calistirabilirsin:" -ForegroundColor Cyan
Write-Host "docker-compose -f docker-compose.prod.yml pull"
Write-Host "docker-compose -f docker-compose.prod.yml up -d"
