#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const yargs = require('yargs');

const options = yargs
    .usage('Usage: -c <config.json> -d <output-directory>')
    .option("c", {alias: "config", describe: "config file path", type: "string", demandOption: true})
    .option("o", {alias: "out", describe: "output directory path", type: "string", demandOption: true})
    .argv;

const configFile = options.config;
const outDir = options.out;

try {
    // Make sure that config file exists
    if (! fs.existsSync(configFile)) {
        throw new Error(`Config file "${configFile}" was not found`);
    }

    // Create output directory (if needed)
    if (!fs.existsSync(outDir)){
        fs.mkdirSync(outDir);
    }

    // Read JSON config file and parse into object
    const configContent = JSON.parse(fs.readFileSync(configFile));

    // Loop over all defined templates in configuration file
    configContent.templates.forEach((templateConfig) => {
        // Make sure that template layout file exists
        if (! fs.existsSync(templateConfig.path)) {
            throw new Error(`Template "${templateConfig.name}" was not found in "${templateConfig.path}"`);
        }

        // Read layout content into memory prepare output directory for template
        const layoutContent = String(fs.readFileSync(templateConfig.path));
        const templateOutDir = path.join(outDir, templateConfig.name);

        if (!fs.existsSync(templateOutDir)){
            fs.mkdirSync(templateOutDir);
        }

        console.info(`Use template "${templateConfig.name}" located in "${templateConfig.path}"`);

        // Loop over all pages
        configContent.pages.forEach((pageConfig) => {
            let outPath = path.join(templateOutDir, `${pageConfig.code}.${configContent.output.file_extension}`);

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
            })
        });
    })
} catch (err) {
    console.error(err);

    process.exit(1);
}
