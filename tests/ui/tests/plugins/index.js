const fs = require('fs');
const os = require('os');
const puppeteer = require('@puppeteer/browsers');

const chromeVersion = "121.0.6167.85"

module.exports = async (on, config) => {
    on('after:spec', (spec, results) => {
        if (results && results.video) {
            // Do we have failures for any retry attempts?
            const failures = results.tests.some((test) =>
                test.attempts.some((attempt) => attempt.state === 'failed')
            )
            if (!failures) {
                // delete the video if the spec passed and no tests retried
                fs.unlinkSync(results.video)
            }
        }
    });

    if (isLinux()) {
        const chromium = await installChromium()
        config.browsers.push(chromium)
    }

    return config;
};

function isLinux() {
    return os.platform() === 'linux'
}

async function installChromium() {
    process.cwd()

    const browser = await puppeteer.install({
        cacheDir: process.cwd() + "/cypress/browsers/chromium",
        browser: "chrome",
        buildId: chromeVersion,
        baseUrl: "https://storage.googleapis.com/chrome-for-testing-public",
    })

    return {
        name: 'chromium',
        family: 'chromium',
        displayName: 'Chromium',
        version: browser.buildId,
        majorVersion: browser.buildId,
        path: browser.executablePath,
        channel: 'ci'
    }
}