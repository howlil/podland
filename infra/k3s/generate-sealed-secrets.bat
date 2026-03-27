@echo off
REM Script untuk generate Sealed Secrets (Windows)
REM Usage: generate-sealed-secrets.bat

setlocal enabledelayedexpansion

set NAMESPACE=podland
set SECRETS_DIR=secrets

echo.
echo Sealed Secrets Generator
echo ==========================
echo.

REM Check kubectl connection
kubectl cluster-info >nul 2>&1
if errorlevel 1 (
    echo Cannot connect to Kubernetes cluster!
    echo Make sure k3s/k3d is running and kubectl is configured
    exit /b 1
)

echo Connected to cluster

REM Check if Sealed Secrets controller is running
kubectl get pods -n kube-system -l name=sealed-secrets-controller >nul 2>&1
if errorlevel 1 (
    echo Sealed Secrets controller not found!
    echo Install with: kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml
    exit /b 1
)

echo Sealed Secrets controller is running
echo.

REM Create secrets directory
if not exist "%SECRETS_DIR%" mkdir "%SECRETS_DIR%"

REM Generate passwords
set /p POSTGRES_PASSWORD="Enter PostgreSQL password (or press Enter to auto-generate): "
if "!POSTGRES_PASSWORD!"=="" (
    for /f "delims=" %%i in ('powershell -Command "-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})"') do set POSTGRES_PASSWORD=%%i
    echo Generated PostgreSQL password: !POSTGRES_PASSWORD!
)

set /p JWT_SECRET="Enter JWT secret (min 32 chars, or press Enter to auto-generate): "
if "!JWT_SECRET!"=="" (
    for /f "delims=" %%i in ('powershell -Command "-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})"') do set JWT_SECRET=%%i
    echo Generated JWT secret: !JWT_SECRET!
)

set /p REFRESH_SECRET="Enter Refresh Token secret (min 32 chars, or press Enter to auto-generate): "
if "!REFRESH_SECRET!"=="" (
    for /f "delims=" %%i in ('powershell -Command "-join ((65..90) + (97..122) + (48..57) | Get-Random -Count 32 | ForEach-Object {[char]$_})"') do set REFRESH_SECRET=%%i
    echo Generated Refresh Token secret: !REFRESH_SECRET!
)

set /p GITHUB_CLIENT_ID="Enter GitHub Client ID: "
set /p GITHUB_CLIENT_SECRET="Enter GitHub Client Secret: "

echo.
echo Creating sealed secrets...

REM Create PostgreSQL sealed secret using kubeseal raw mode
echo -n !POSTGRES_PASSWORD! | kubeseal --raw --from-file=/dev/stdin --namespace %NAMESPACE% --name postgres-secret --key password > /tmp/postgres-password.encrypted
set /p ENCRYPTED_PASSWORD=< /tmp/postgres-password.encrypted

echo -n !JWT_SECRET! | kubeseal --raw --from-file=/dev/stdin --namespace %NAMESPACE% --name podland-backend-secret --key jwt-secret > /tmp/jwt-secret.encrypted
set /p ENCRYPTED_JWT=< /tmp/jwt-secret.encrypted

echo -n !REFRESH_SECRET! | kubeseal --raw --from-file=/dev/stdin --namespace %NAMESPACE% --name podland-backend-secret --key refresh-token-secret > /tmp/refresh-secret.encrypted
set /p ENCRYPTED_REFRESH=< /tmp/refresh-secret.encrypted

echo -n !GITHUB_CLIENT_ID! | kubeseal --raw --from-file=/dev/stdin --namespace %NAMESPACE% --name podland-backend-secret --key github-client-id > /tmp/github-id.encrypted
set /p ENCRYPTED_GITHUB_ID=< /tmp/github-id.encrypted

echo -n !GITHUB_CLIENT_SECRET! | kubeseal --raw --from-file=/dev/stdin --namespace %NAMESPACE% --name podland-backend-secret --key github-client-secret > /tmp/github-secret.encrypted
set /p ENCRYPTED_GITHUB_SECRET=< /tmp/github-secret.encrypted

REM Create SealedSecret YAML files
(
echo apiVersion: bitnami.com/v1alpha1
echo kind: SealedSecret
echo metadata:
echo   name: postgres-secret
echo   namespace: %NAMESPACE%
echo spec:
echo   encryptedData:
echo     password: !ENCRYPTED_PASSWORD!
echo   template:
echo     metadata:
echo       name: postgres-secret
echo       namespace: %NAMESPACE%
echo     type: Opaque
) > "%SECRETS_DIR%\postgres-sealedsecret.yaml"

(
echo apiVersion: bitnami.com/v1alpha1
echo kind: SealedSecret
echo metadata:
echo   name: podland-backend-secret
echo   namespace: %NAMESPACE%
echo spec:
echo   encryptedData:
echo     jwt-secret: !ENCRYPTED_JWT!
echo     refresh-token-secret: !ENCRYPTED_REFRESH!
echo     github-client-id: !ENCRYPTED_GITHUB_ID!
echo     github-client-secret: !ENCRYPTED_GITHUB_SECRET!
echo   template:
echo     metadata:
echo       name: podland-backend-secret
echo       namespace: %NAMESPACE%
echo     type: Opaque
) > "%SECRETS_DIR%\backend-sealedsecret.yaml"

REM Cleanup
del /f /q /tmp\*.encrypted 2>nul

echo.
echo Sealed secrets generated successfully!
echo.
echo Files created:
echo    - %SECRETS_DIR%\postgres-sealedsecret.yaml
echo    - %SECRETS_DIR%\backend-sealedsecret.yaml
echo.
echo Next steps:
echo    1. Review the sealed secret files
echo    2. Apply to cluster: kubectl apply -f %SECRETS_DIR%\
echo    3. Commit sealed secrets to git ^(safe to commit!^)
echo.
echo IMPORTANT: Save these values somewhere safe!
echo    PostgreSQL Password: !POSTGRES_PASSWORD!
echo    JWT Secret: !JWT_SECRET!
echo    Refresh Token Secret: !REFRESH_SECRET!

endlocal
