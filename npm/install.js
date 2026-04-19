const fs = require('fs');
const path = require('path');
const https = require('https');
const os = require('os');

const VERSION = '1.2.4';
const REPO = 'Bekican/gorenel';

function getBinaryName() {
    const platform = os.platform();
    const arch = os.arch();

    const platMap = {
        'darwin': 'macos',
        'linux': 'linux',
        'win32': 'windows'
    };

    const archMap = {
        'arm64': 'arm64',
        'x64': 'x64'
    };

    const p = platMap[platform];
    const a = archMap[arch];

    if (!p || !a) {
        throw new Error(`Unsupported platform: ${platform} ${arch}`);
    }

    const ext = p === 'windows' ? '.exe' : '';
    return `gorenel-${p}-${a}${ext}`;
}

async function download() {
    const binaryName = getBinaryName();
    const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;
    const destDir = path.join(__dirname, 'bin');
    const destPath = path.join(destDir, binaryName === getBinaryName() ? (os.platform() === 'win32' ? 'gorenel.exe' : 'gorenel') : 'gorenel');

    if (!fs.existsSync(destDir)) {
        fs.mkdirSync(destDir);
    }

    console.log(`⚡ Downloading Gorenel CLI v${VERSION} for ${os.platform()} ${os.arch()}...`);
    console.log(`🔗 Source: ${url}`);

    const file = fs.createWriteStream(destPath);

    https.get(url, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
            // Handle redirect (GitHub Releases redirects to S3/Azure)
            https.get(response.headers.location, (redirectResponse) => {
                redirectResponse.pipe(file);
            });
        } else if (response.statusCode !== 200) {
            console.error(`❌ Failed to download binary. Status: ${response.statusCode}`);
            console.error(`Please ensure v${VERSION} is released on GitHub.`);
            process.exit(1);
        } else {
            response.pipe(file);
        }

        file.on('finish', () => {
            file.close();
            if (os.platform() !== 'win32') {
                fs.chmodSync(destPath, 0o755);
            }
            console.log('✅ Gorenel CLI installed successfully!');
        });
    }).on('error', (err) => {
        fs.unlink(destPath, () => {});
        console.error(`❌ Error downloading binary: ${err.message}`);
        process.exit(1);
    });
}

// Only run if not being required
if (require.main === module) {
    download().catch(err => {
        console.error(err);
        process.exit(1);
    });
}
