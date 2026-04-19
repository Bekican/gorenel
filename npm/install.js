const fs = require('fs');
const path = require('path');
const os = require('os');

const VERSION = '1.2.5';

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

async function setup() {
    const binaryName = getBinaryName();
    const binDir = path.join(__dirname, 'bin');
    const binPath = path.join(binDir, binaryName);

    console.log(`⚡ Setting up Gorenel CLI v${VERSION} for ${os.platform()} ${os.arch()}...`);

    if (!fs.existsSync(binPath)) {
        console.error(`❌ Binary not found for your platform: ${binaryName}`);
        console.error(`Check the 'npm/bin' directory for bundled files.`);
        process.exit(1);
    }

    // Set executable permissions for Unix/Mac
    if (os.platform() !== 'win32') {
        try {
            fs.chmodSync(binPath, 0o755);
            console.log('✅ Executable permissions set.');
        } catch (err) {
            console.warn(`⚠️ Failed to set permissions: ${err.message}`);
        }
    }

    console.log('✅ Gorenel CLI setup complete!');
}

if (require.main === module) {
    setup().catch(err => {
        console.error(err);
        process.exit(1);
    });
}
