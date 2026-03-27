@echo off
setlocal enabledelayedexpansion

:: Phase 2 Interactive Discussion CLI
:: Usage: 2-discuss.bat

echo.
echo ============================================
echo  Phase 2: Core VM - Interactive Discussion
echo ============================================
echo.

:: Question 1: VM Image Strategy - Base Image
:Q1
echo [Question 1 of 4] VM Image Strategy - Base Image
echo.
echo Context: Kita sudah putuskan pakai pre-built images (ubuntu-2204, debian-12).
echo Sekarang perlu detail: base image dari mana?
echo.
echo Options:
echo   [A] Official Docker Hub images (ubuntu:22.04, debian:12)
echo       + Trusted source, regularly updated
echo       + Minimal attack surface
echo       - Very minimal (no common tools)
echo       - User perlu install everything
echo.
echo   [B] Build sendiri dari official images (layer on top)
echo       + Add common tools (git, curl, wget)
echo       + Consistent environment across VMs
echo       + Can optimize for size
echo       - Maintenance overhead (rebuild when base updates)
echo.
echo   [C] Third-party images (linuxserver.io, etc.)
echo       + Feature-rich
echo       - Less control over what's included
echo       - Trust dependency
echo.
set /p Q1_ANSWER="Your choice (A/B/C): "
if /i "%Q1_ANSWER%"=="A" set Q1_RESULT=Official Docker Hub images
if /i "%Q1_ANSWER%"=="B" set Q1_RESULT=Build sendiri dari official images
if /i "%Q1_ANSWER%"=="C" set Q1_RESULT=Third-party images
echo.
echo Selected: %Q1_RESULT%
echo.
pause
echo.

:: Question 2: VM Image Strategy - Preinstall Packages
:Q2
echo [Question 2 of 4] VM Image Strategy - Preinstall Packages
echo.
echo Context: Base image sudah dipilih. Sekarang apa yang di-preinstall?
echo.
echo Options:
echo   [A] Essentials only (recommended)
echo       Packages: git, curl, wget, vim, htop, net-tools
echo       + Fast image pull (small size)
echo       + Minimal attack surface
echo       - User install language runtimes separately
echo.
echo   [B] Essentials + Build tools
echo       Packages: essentials + gcc, make, build-essential
echo       + User can compile from source
echo       - Larger image (500MB+)
echo       - Longer build time
echo.
echo   [C] Full-featured (essentials + Node.js + Python + Go)
echo       + Ready to code immediately
echo       - Very large image (1GB+)
echo       - Security updates for all packages
echo.
set /p Q2_ANSWER="Your choice (A/B/C): "
if /i "%Q2_ANSWER%"=="A" set Q2_RESULT=Essentials only (git, curl, wget, vim, htop, net-tools)
if /i "%Q2_ANSWER%"=="B" set Q2_RESULT=Essentials + Build tools (gcc, make, build-essential)
if /i "%Q2_ANSWER%"=="C" set Q2_RESULT=Full-featured (essentials + Node.js + Python + Go)
echo.
echo Selected: %Q2_RESULT%
echo.
pause
echo.

:: Question 3: VM Image Strategy - Build & Hosting
:Q3
echo [Question 3 of 4] VM Image Strategy - Build ^& Hosting
echo.
echo Context: Image content sudah jelas. Sekarang build ^& hosting strategy?
echo.
echo Options:
echo   [A] GitHub Actions + Docker Hub (recommended)
echo       + Automated monthly builds
echo       + Free for open source
echo       + Versioned images (v1, v2, etc.)
echo       - Public repo (or private with limits)
echo.
echo   [B] Self-hosted registry (k3s)
echo       + Full control
echo       + Private by default
echo       - Need to manage registry
echo       - Storage on same server
echo.
echo   [C] Manual build + Docker Hub
echo       + Full control over timing
echo       - Manual process (error-prone)
echo       - No automation
echo.
set /p Q3_ANSWER="Your choice (A/B/C): "
if /i "%Q3_ANSWER%"=="A" set Q3_RESULT=GitHub Actions + Docker Hub
if /i "%Q3_ANSWER%"=="B" set Q3_RESULT=Self-hosted registry (k3s)
if /i "%Q3_ANSWER%"=="C" set Q3_RESULT=Manual build + Docker Hub
echo.
echo Selected: %Q3_RESULT%
echo.
pause
echo.

:: Question 4: VM Image Strategy - Versioning & Updates
:Q4
echo [Question 4 of 4] VM Image Strategy - Versioning ^& Updates
echo.
echo Context: Image sudah di-build ^& hosted. Sekarang versioning strategy?
echo.
echo Options:
echo   [A] Semantic versioning + monthly builds (recommended)
echo       Tags: v1.0, v1.1, 2026-03, latest
echo       + Clear version history
echo       + Monthly security patches
echo       + User can pin to specific version
echo       - Multiple tags to maintain
echo.
echo   [B] Latest only (rolling update)
echo       Tag: latest
echo       + Simple
echo       - No reproducibility
echo       - Breaking changes possible
echo.
echo   [C] Date-based only
echo       Tags: 2026-03-27, 2026-04-01
echo       + Clear when image was built
echo       - No semantic meaning
echo.
set /p Q4_ANSWER="Your choice (A/B/C): "
if /i "%Q4_ANSWER%"=="A" set Q4_RESULT=Semantic versioning + monthly builds
if /i "%Q4_ANSWER%"=="B" set Q4_RESULT=Latest only (rolling)
if /i "%Q4_ANSWER%"=="C" set Q4_RESULT=Date-based only
echo.
echo Selected: %Q4_RESULT%
echo.
pause
echo.

:: Summary
echo ============================================
echo  Summary: VM Image Strategy Decisions
echo ============================================
echo.
echo Q1 Base Image:        %Q1_RESULT%
echo Q2 Preinstall:        %Q2_RESULT%
echo Q3 Build ^& Hosting:   %Q3_RESULT%
echo Q4 Versioning:        %Q4_RESULT%
echo.
echo Saving to: .planning/phases/2-CONTEXT-updates.md
echo.

:: Save decisions
(
echo # Phase 2 Discussion Results
echo.
echo ## VM Image Strategy
echo.
echo - **Base Image:** %Q1_RESULT%
echo - **Preinstall:** %Q2_RESULT%
echo - **Build ^& Hosting:** %Q3_RESULT%
echo - **Versioning:** %Q4_RESULT%
echo.
) > .planning\phases\2-CONTEXT-updates.md

echo Done! Results saved.
echo.
echo Next: Continue to next gray area? (Y/N)
set /p CONTINUE="Continue: "
if /i "%CONTINUE%"=="Y" goto:GRAY2
echo.
echo Discussion session ended.
echo Run this script again to continue other gray areas.
pause
exit /b

:: Gray Area 2: VM Startup & Shutdown Flow
:GRAY2
echo.
echo ============================================
echo  Gray Area 2: VM Startup ^& Shutdown Flow
echo ============================================
echo.
echo [Loading next questions...]
echo.
:: (Continue with next gray area questions...)
pause
