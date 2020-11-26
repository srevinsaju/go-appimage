#!/usr/bin/env bash
set -eux

mv "./$BUILD_APP" ./$BUILD_APP-$(go env GOHOSTARCH)
# export the ARCHITECTURE
export ARCHITECTURE=$BUIILD_ARCH
if [[ "$BUIILD_ARCH" == "386" ]]; then
    export ARCHITECTURE="i686"
fi
if [[ "$BUIILD_ARCH" == "amd64" ]]; then
    export ARCHITECTURE="x86_64"
fi

mkdir -p "$BUILD_APP.AppDir/usr/bin"

if [[ "$BUILD_APP" != "appimaged" ]]; then
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/static-tools/releases/download/continuous/desktop-file-validate-$ARCHITECTURE -O desktop-file-validate )
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/static-tools/releases/download/continuous/mksquashfs-$ARCHITECTURE -O mksquashfs )
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/static-tools/releases/download/continuous/patchelf-$ARCHITECTURE -O patchelf )
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/AppImage/AppImageKit/releases/download/continuous/runtime-$ARCHITECTURE )
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/uploadtool/raw/master/upload.sh -O uploadtool )
fi
if [[ "$BUILD_APP" != "appimagetool" ]]; then
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/static-tools/releases/download/continuous/bsdtar-$ARCHITECTURE -O bsdtar )
    ( cd "$BUILD_APP.AppDir/usr/bin/" ; wget -c https://github.com/probonopd/static-tools/releases/download/continuous/unsquashfs-$ARCHITECTURE -O unsquashfs )
fi
chmod +x $BUILD_APP.AppDir/usr/bin/*
cp "$BUILD_APP-$(go env GOHOSTARCH)" "$BUILD_APP.AppDir/usr/bin/$BUILD_APP"
( cd $BUILD_APP.AppDir/ ; ln -s usr/bin/$BUILD_APP AppRun)
cp data/appimage.png $BUILD_APP.AppDir/
cat > $BUILD_APP.AppDir/$BUILD_APP.desktop <<\EOF
[Desktop Entry]
Type=Application
Name=$BUILD_APP
Exec=$BUILD_APP
Comment=$BUILD_APP - tool to generate AppImages from AppDirs
Icon=appimage
Categories=Development;
Terminal=true
EOF
if [[ "$BUILD_APP" == "appimagetool" ]]; then
    ln -s $BUILD_APP.AppDir/usr/bin/* .
    PATH="$BUILD_APP.AppDir/usr/bin/:$PATH" ./appimagetool-* ./${{ matrix.app}}.AppDir || true  # FIXME: remove this true
else
    # use our own dog food :)
    chmod +x ./appimagetool-*-deploy*.AppImage/*.AppImage
    ./appimagetool-*.AppImage/*.AppImage ./${{ matrix.app}}.AppDir || true
fi
rm -rf ./appimagetool-*-deploy*.AppImage
mkdir dist
mv *.AppImage dist
