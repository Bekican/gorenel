#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const os = require('os');

const binaryName = os.platform() === 'win32' ? 'gorenel.exe' : 'gorenel';
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
        console.error('❌ Gorenel binary not found. Please try reinstalling: npm install -g gorenel');
    } else {
        console.error(`❌ Error executing Gorenel: ${err.message}`);
    }
    process.exit(1);
});
