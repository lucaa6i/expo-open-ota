const fs = require("fs");
const path = require("path");

const manifestPath = path.join(__dirname, "../android/app/src/main/AndroidManifest.xml");
const oldUrl = "http://localhost:3000";
const newUrl = "http://10.0.2.2:3000";

if (!fs.existsSync(manifestPath)) {
  console.error("❌ AndroidManifest.xml not found");
  process.exit(1);
}

let content = fs.readFileSync(manifestPath, "utf8");
if (!content.includes(oldUrl)) {
  console.log("ℹ️ No update needed");
  process.exit(0);
}

content = content.replace(new RegExp(oldUrl, "g"), newUrl);
fs.writeFileSync(manifestPath, content, "utf8");
