const express = require("express");
const puppeteer = require("puppeteer-core");
const chromium = require("@sparticuz/chromium");

const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3001;

app.post("/render", async (req, res) => {
    const { url } = req.body;
    if (!url) return res.status(400).send("Missing URL");
    let browser;
    try {
        browser = await puppeteer.launch({
            args: chromium.args,
            executablePath: await chromium.executablePath(),
            headless: chromium.headless,
        });
        const page = await browser.newPage();
        await page.goto(url, { waitUntil: "networkidle2", timeout: 30000 });
        const html = await page.content();
        await browser.close();

        res.setHeader("Content-Type", "text/html");
        res.send(html);
    } catch (err) {
        console.error("Puppeteer render error:", err);
        if (browser) await browser.close();
        res.status(500).send("Failed to render page");
    }
});

app.listen(PORT, "0.0.0.0", () => {
    console.log(`Render server running on http://0.0.0.0:${PORT}`);
});