#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building dpkg package for duofm...${NC}"

# Get project information
PROJECT_NAME="duofm"
VERSION=$(git describe --tags --always 2>/dev/null | sed 's/^v//' || echo "0.1.0")
ARCH=$(uname -m)
MAINTAINER="m-m-n <51132276+m-m-n@users.noreply.github.com>"

# Convert architecture to Debian format
case "$ARCH" in
    x86_64)
        DEB_ARCH="amd64"
        ;;
    aarch64)
        DEB_ARCH="arm64"
        ;;
    armv7l)
        DEB_ARCH="armhf"
        ;;
    i686)
        DEB_ARCH="i386"
        ;;
    *)
        DEB_ARCH="$ARCH"
        ;;
esac

PACKAGE_NAME="${PROJECT_NAME}_${VERSION}_${DEB_ARCH}"
BUILD_DIR="build/dpkg/${PACKAGE_NAME}"

echo ""
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo -e "${YELLOW}Package: ${PACKAGE_NAME}${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Architecture: ${DEB_ARCH}${NC}"
echo -e "${YELLOW}Maintainer: ${MAINTAINER}${NC}"
echo -e "${BLUE}═══════════════════════════════════════${NC}"
echo ""

# Check if dpkg-deb is available
if ! command -v dpkg-deb &> /dev/null; then
    echo -e "${RED}Error: dpkg-deb command not found${NC}"
    echo "Please install dpkg tools: sudo apt-get install dpkg"
    exit 1
fi

# Clean previous build
if [ -d "build/dpkg" ]; then
    echo "Cleaning previous build..."
    rm -rf build/dpkg
fi

# Create directory structure
echo "Creating package directory structure..."
mkdir -p "${BUILD_DIR}/DEBIAN"
mkdir -p "${BUILD_DIR}/usr/bin"
mkdir -p "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}"
mkdir -p "${BUILD_DIR}/usr/share/man/man1"

# Build the binary
echo "Building Go binary..."
if ! make build; then
    echo -e "${RED}Failed to build binary${NC}"
    exit 1
fi

# Verify binary exists
if [ ! -f "${PROJECT_NAME}" ]; then
    echo -e "${RED}Error: Binary ${PROJECT_NAME} not found after build${NC}"
    exit 1
fi

# Copy binary
echo "Copying binary to package..."
cp "${PROJECT_NAME}" "${BUILD_DIR}/usr/bin/"
chmod 755 "${BUILD_DIR}/usr/bin/${PROJECT_NAME}"

# Copy documentation
echo "Copying documentation..."
if [ -f "README.md" ]; then
    cp README.md "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/"
    chmod 644 "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/README.md"
fi

if [ -f "LICENSE" ]; then
    cp LICENSE "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/copyright"
    chmod 644 "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/copyright"
elif [ -f "LICENCE" ]; then
    cp LICENCE "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/copyright"
    chmod 644 "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/copyright"
fi

# Create changelog
echo "Creating changelog..."
cat > "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/changelog" << EOF
${PROJECT_NAME} (${VERSION}) stable; urgency=low

  * Release version ${VERSION}
  * See git history for detailed changes

 -- ${MAINTAINER}  $(date -R)
EOF
chmod 644 "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/changelog"

# Compress changelog
if command -v gzip &> /dev/null; then
    gzip -9 "${BUILD_DIR}/usr/share/doc/${PROJECT_NAME}/changelog"
fi

# Create DEBIAN/control file
echo "Creating control file..."
cat > "${BUILD_DIR}/DEBIAN/control" << 'EOF'
Package: duofm
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${DEB_ARCH}
Maintainer: m-m-n <51132276+m-m-n@users.noreply.github.com>
Depends: libc6
Description: Unifies Orthodox File Manipulation
 A terminal-based dual-pane file manager written in Go, inspired by
 classic file managers with vim-style keybindings.
 .
 Features:
  - Dual-pane interface for easy file navigation
  - Vim-style keybindings (hjkl)
  - File operations (copy, move, delete)
  - Modal dialogs with confirmation
  - Built-in help system
  - Cross-platform terminal application
EOF

# Substitute variables in control file
sed -i "s/\${VERSION}/${VERSION}/g" "${BUILD_DIR}/DEBIAN/control"
sed -i "s/\${DEB_ARCH}/${DEB_ARCH}/g" "${BUILD_DIR}/DEBIAN/control"

# Create postinst script
echo "Creating postinst script..."
cat > "${BUILD_DIR}/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Ensure binary is executable
if [ -f /usr/bin/duofm ]; then
    chmod 755 /usr/bin/duofm
fi

echo "duofm installed successfully!"
echo "Run 'duofm' to start the file manager."

exit 0
EOF
chmod 755 "${BUILD_DIR}/DEBIAN/postinst"

# Create prerm script
echo "Creating prerm script..."
cat > "${BUILD_DIR}/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

# Clean up before removal (if needed)

exit 0
EOF
chmod 755 "${BUILD_DIR}/DEBIAN/prerm"

# Create postrm script
echo "Creating postrm script..."
cat > "${BUILD_DIR}/DEBIAN/postrm" << 'EOF'
#!/bin/bash
set -e

# Clean up after removal
echo "duofm has been removed."

exit 0
EOF
chmod 755 "${BUILD_DIR}/DEBIAN/postrm"

# Set proper permissions
echo "Setting file permissions..."
find "${BUILD_DIR}/usr/share/doc" -type f -exec chmod 644 {} \;
find "${BUILD_DIR}/usr/share/doc" -type d -exec chmod 755 {} \;

# Calculate installed size (in KB)
INSTALLED_SIZE=$(du -sk "${BUILD_DIR}" | cut -f1)
echo "Installed-Size: ${INSTALLED_SIZE}" >> "${BUILD_DIR}/DEBIAN/control"

# Build the package
echo ""
echo -e "${BLUE}Building .deb package...${NC}"
if dpkg-deb --build --root-owner-group "${BUILD_DIR}"; then
    # Move the package to the root directory
    mv "build/dpkg/${PACKAGE_NAME}.deb" .

    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}✓ Package created successfully!${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo -e "${BLUE}Package file: ${YELLOW}${PACKAGE_NAME}.deb${NC}"

    # Show package info
    echo ""
    echo -e "${BLUE}Package Information:${NC}"
    dpkg-deb --info "${PACKAGE_NAME}.deb" | head -20

    echo ""
    echo -e "${BLUE}Package Contents:${NC}"
    dpkg-deb --contents "${PACKAGE_NAME}.deb"

    echo ""
    echo -e "${GREEN}Installation Commands:${NC}"
    echo -e "  ${YELLOW}sudo dpkg -i ${PACKAGE_NAME}.deb${NC}     - Install package"
    echo -e "  ${YELLOW}dpkg -L duofm${NC}                        - List installed files"
    echo -e "  ${YELLOW}sudo dpkg -r duofm${NC}                   - Remove package"
    echo -e "  ${YELLOW}sudo dpkg -P duofm${NC}                   - Purge package completely"
    echo ""
else
    echo -e "${RED}Failed to build package${NC}"
    exit 1
fi

# Clean up build directory
echo "Cleaning up build directory..."
rm -rf build/dpkg

echo -e "${GREEN}Done!${NC}"
