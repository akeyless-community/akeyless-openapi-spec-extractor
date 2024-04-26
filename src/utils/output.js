// utils/output.js
import yaml from 'js-yaml';

export const outputResult = (data, format) => {
  const outputFormat = format.toLowerCase();
  let outputData;

  if (outputFormat === 'json') {
    outputData = JSON.stringify(data, null, 2);
  } else if (outputFormat === 'yaml') {
    outputData = yaml.dump(data);
  } else {
    throw new Error(`Unsupported output format: ${outputFormat}`);
  }

  console.log(outputData);
};