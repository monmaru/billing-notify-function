'use strict';

const _ = require('lodash');
const rp = require('request-promise');
const storage = require('@google-cloud/storage')();

// [START your slack setting]
const WEBHOOK_URL = '';
const BOT_NAME ='gcp-billing-bot';
// [END your slack setting]

function getFileStream (file) {
  if (!file.bucket) {
    throw new Error('Bucket not provided. Make sure you have a "bucket" property in your request');
  }
  if (!file.name) {
    throw new Error('Filename not provided. Make sure you have a "name" property in your request');
  }

  return storage.bucket(file.bucket).file(file.name).createReadStream();
}

function post2Slack (fileName, billing) {
  const fields = _.map(billing, (v) => {
    return {
      title: `${v.projectId}: ${v.description}`,
      value: `${v.cost.amount}ドル（USD）`
    }
  });

  const requestBody = {
    username: BOT_NAME,
    pretext: fileName.match(/billing-(.*).json/)[1] + 'の請求書',
    color: '#36a64f',
    fields: fields
  }

  const params = {
    method: 'POST',
    uri: WEBHOOK_URL,
    body: requestBody,
    json: true
  };
  return rp(params);
}

exports.notifyBillingInfo = function notifyBillingInfo (event) {
  const file = event.data;
  return Promise.resolve()
    .then(() => {
      if (file.resourceState === 'not_exists') {
        return;
      }

      let text = '';
      getFileStream(file).on('data', (chunk) => {
        text += chunk;
      }).on('end', () => {
        return post2Slack(file.name, JSON.parse(text));
      });
    })
    .then(() => {
      console.log(`File ${file.name} processed.`);
    })
    .catch((err) => {
      console.error(err);
      return err;
    });
};
