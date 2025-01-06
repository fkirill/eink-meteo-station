const puppeteer = require('puppeteer');

const delay = (millis) => {
    var resolver;
    const promise = new Promise((resolverFn) => {
        resolver = resolverFn;
    });
    setTimeout(() => {
        resolver();
    }, millis);
    return promise;
}

(async () => {
    const cmdlineArguments = process.argv.slice(2);
    const sourceHtmlFile = cmdlineArguments[0];
    const targetPngFile = cmdlineArguments[1];
    const viewportWidth = parseInt(cmdlineArguments[2]);
    const viewportHeight = parseInt(cmdlineArguments[3]);

    const browser = await puppeteer.launch({
        headless: true,
        executablePath: '/Applications/Chromium.app/Contents/MacOS/Chromium',
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    const page = await browser.newPage();
    await page.setViewport({ width: viewportWidth, height: viewportHeight });
    await page.goto("file://" + sourceHtmlFile);
    await delay(1000);
    await page.screenshot({ path: targetPngFile });
    await browser.close();
})();











puppeteer = require('puppeteer');

browserPromise = puppeteer.launch({
    headless: true,
    executablePath: '/usr/bin/chromium',
    args: ['--no-sandbox', '--disable-setuid-sandbox']
});

(async () => {
    const browser = await browserPromise;
    const page = await browser.newPage();
    await page.setViewport({ width: 962, height: 1400 });
    await page.goto("file:///home/pi/eink-meteo-station/temp.html");
    await page.waitForNetworkIdle({options:{idleTime:0}});
    await page.screenshot({ path: "/home/pi/eink-meteo-station/temp.png" });
    await page.close();
    console.log("Done")
})();





(async (browserPromise, htmlContent, viewportWidth, viewportHeight) => {
    const browser = await browserPromise;
    const page = await browser.newPage();
    await page.setViewport({ width: viewportWidth, height: viewportHeight });
    await page.setContent(htmlContent, {options: {waitUntil:"networkidle0"}});
    const buf = await page.screenshot({ encoding: "base64", type: "png"});
    await page.close();
    console.log(buf);
})(browserPromise, '%s', 1, 2);