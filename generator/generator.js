#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const yargs = require('yargs');

const options = yargs
    .usage('Usage: -c <config.json> -d <output-directory>')
    .option("c", {alias: "config", describe: "config file path", type: "string", demandOption: true})
    .option("o", {alias: "out", describe: "output directory path", type: "string", demandOption: true})
    .option("i", {alias: "index", describe: "generate index page", type: "boolean"})
    .argv;

const configFile = options.config;
const outDir = options.out;
const generateIndexPage = options.index;

const generated = {};

try {
    // Make sure that config file exists
    if (!fs.existsSync(configFile)) {
        throw new Error(`Config file "${configFile}" was not found`);
    }

    // Create output directory (if needed)
    if (!fs.existsSync(outDir)) {
        fs.mkdirSync(outDir);
    }

    // Read JSON config file and parse into object
    const configContent = JSON.parse(fs.readFileSync(configFile));

    // Loop over all defined templates in configuration file
    configContent.templates.forEach((templateConfig) => {
        // Make sure that template layout file exists
        if (!fs.existsSync(templateConfig.path)) {
            throw new Error(`Template "${templateConfig.name}" was not found in "${templateConfig.path}"`);
        }

        // Read layout content into memory prepare output directory for template
        const layoutContent = String(fs.readFileSync(templateConfig.path));
        const templateOutDir = path.join(outDir, templateConfig.name);

        if (!fs.existsSync(templateOutDir)) {
            fs.mkdirSync(templateOutDir);
        }

        console.info(`Use template "${templateConfig.name}" located in "${templateConfig.path}"`);

        // Loop over all pages
        configContent.pages.forEach((pageConfig) => {
            let outFileName = pageConfig.code + "." + configContent.output.file_extension,
                outPath = path.join(templateOutDir, outFileName);

            console.info(`  [${templateConfig.name}:${pageConfig.code}] Output: ${outPath}`);

            // Make replaces
            let result = layoutContent
                .replace(/{{\s?code\s?}}/g, pageConfig.code)
                .replace(/{{\s?message\s?}}/g, pageConfig.message)
                .replace(/{{\s?description\s?}}/g, pageConfig.description);

            // And write into result file
            fs.writeFileSync(outPath, result, {
                encoding: "utf8",
                flag: "w+",
                mode: 0o644
            });

            if (!generated[templateConfig.name]) {
                generated[templateConfig.name] = [];
            }

            generated[templateConfig.name].push({
                code: pageConfig.code,
                message: pageConfig.message,
                description: pageConfig.description,
                path: path.join(templateConfig.name, outFileName),
            })
        });
    })

    // Generate index page for the generated content
    if (generateIndexPage === true) {
        const indexPageFilePath = path.join(outDir, 'index.html');

        let content = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
  <title>Error pages list</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.6.0/css/bootstrap.min.css"
    integrity="sha512-P5MgMn1jBN01asBgU0z60Qk4QxiXo86+wlFahKrsQf37c9cro517WzVSPPV1tDKzhku2iJ2FVgL67wG03SGnNA=="
    crossorigin="anonymous" />
</head>
<body>
<main role="main" class="container">\n`;

        Object.keys(generated).forEach(function(templateName) {
            content += `<h2 class="mt-5">Template name: <code>` + templateName + `</code></h2>\n<ul class="mb-5">\n`;

            generated[templateName].forEach((properties) => {
                content += `  <li><a href="${properties.path}"><span class="badge badge-light">${properties.code}</span>: ${properties.message}</a></li>\n`;
            })

            content += `</ul>\n`;
        });

        content += `</main>
<footer class="footer">
  <div class="container text-center text-muted mt-3 mb-3">
    For online documentation and support please refer to the <a href="https://github.com/tarampampam/error-pages">project repository</a>.
  </div>
</footer>
</body>
</html>`;

        fs.writeFileSync(indexPageFilePath, content, {
            encoding: "utf8",
            flag: "w+",
            mode: 0o644
        });
    }
} catch (err) {
    console.error(err);

    process.exit(1);
}
