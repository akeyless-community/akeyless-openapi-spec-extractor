// commands/fetch.js
import { Command } from 'commander';
import openapiExtractor from 'openapi-extract';
import axios from 'axios';
import { logger } from '../utils/logger.js';
import { outputResult } from '../utils/output.js';

export const fetchCommand = new Command('fetch')
  .description('Fetch and process OpenAPI spec')
  .requiredOption('-u, --url <url>', 'URL of the OpenAPI spec')
  .requiredOption('-p, --path <path>', 'The Path to the desired endpoint within the OpenAPI spec')
  .option('-o, --output <output>', 'Output type (default: "json"): "json" or "yaml"', 'json')
  .option('-l, --loglevel <level>', 'Logging level (default: "error"): "error", "warn", "info", "debug"', 'error')
  .option('-v, --validate', 'Enable validation of OpenAPI spec', false)
  .action(async (options) => {
    try {
      logger.level = options.loglevel;
      logger.info(`Fetching OpenAPI spec from URL: ${options.url}`);

      const response = await axios.get(options.url);
      const openapiSpec = response.data;

      logger.debug('OpenAPI spec fetched successfully');

      // prefix a leading slash to the options.path if it doesn't already have one
      options.path = options.path.startsWith('/') ? options.path : `/${options.path}`;

      const extractedData = openapiExtractor.extract(openapiSpec, { path: options.path, removeDocs: true, removeExamples: true, removeExtensions: true, openai: true });

      logger.debug('Extracted data:', extractedData);

      outputResult(extractedData, options.output);

      logger.info('Extraction completed successfully');
    } catch (error) {
      logger.error('An error occurred:', error);
      process.exit(1);
    }
  });