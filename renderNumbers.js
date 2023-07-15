const puppeteer = require("puppeteer");

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.setViewport({ width: 1872, height: 1404, deviceScaleFactor: 1 });
  await page.goto("file:///Users/kirillfrolov/eink-meteo-station/time.html");
  const minutes = await page.$("#minutes");
  const minutesBox = await minutes.boundingBox();
  const seconds = await page.$("#seconds");
  const secondsBox = await seconds.boundingBox();
  const colon = await page.$("#colon");
  const colonBox = await colon.boundingBox();
  const hours = await page.$("#hours");
  const hoursBox = await hours.boundingBox();
  const boxes = { minutesBox, secondsBox, colonBox, hoursBox };
  console.log("Bounding box", JSON.stringify(boxes));

  for (let i = 0; i < 60; i++) {
    const value = (i < 10 ? "0" : "") + i;
    await minutes.evaluate(
      (elem, [value]) => {
        elem.innerHTML = value;
      },
      [value]
    );
    await minutes.screenshot({ path: "numbers/num_" + value + ".png" });
    console.log("minutes", value);
  }
  await colon.screenshot({ path: "numbers/colon.png" });

  await browser.close();

})();
