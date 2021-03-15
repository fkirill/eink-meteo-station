const puppeteer = require('puppeteer');

const delay = (millis) => {
    var resolver;
    var promise = new Promise((resolverFn) => {
        resolver = resolverFn;
    });
    setTimeout(() => {
        resolver();
    }, millis);
    return promise;
}

(async () => {
    const cmdlineArguments = process.argv.slice(2);
    const sourceHtmlFile = cmdlineArguments[0]
    const targetPngFile = cmdlineArguments[1]
    const viewportWidth = parseInt(cmdlineArguments[2])
    const viewportHeight = parseInt(cmdlineArguments[3])

    const browser = await puppeteer.launch({
        headless: true,
        executablePath: '/usr/bin/chromium-browser',
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    const page = await browser.newPage();
    await page.setViewport({width: viewportWidth, height: viewportHeight})
    await page.goto("file://" + sourceHtmlFile);
    await delay(1000);
    await page.screenshot({path: targetPngFile});
    await browser.close();
})();