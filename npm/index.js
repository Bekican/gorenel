#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const os = require('os');

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

const binaryName = getBinaryName();
const binaryPath = path.join(__dirname, 'bin', binaryName);

const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
    windowsHide: true
});

child.on('close', (code) => {
    process.exit(code);
});

child.on('error', (err) => {
    if (err.code === 'ENOENT') {
        console.error(`❌ Gorenel binary not found: ${binaryName}`);
        console.error('Please try reinstalling: npm install -g gorenel');
    } else {
        console.error(`❌ Error executing Gorenel: ${err.message}`);
    }
    process.exit(1);
});
