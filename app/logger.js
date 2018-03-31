/*
 * Daniele Bissoli
 * FdP Tickets Manager - Logger module
 * v0.0.2 - 2017-04-16
 */

// library for logging
const winston = require("winston");

// logging initialization
module.exports = new (winston.Logger)({
  // transports configuration to choose on
  // which files write and which levels use
  transports: [
    // transport for information log
    new (winston.transports.File)({
      name: 'info-file',
      filename: 'logs/info.log',
      level: 'info'
    }),
    // transport for errors log
    new (winston.transports.File)({
      name: 'error-file',
      filename: 'logs/errors.log',
      level: 'error'
    }),
    // transport for warns log
    new (winston.transports.File)({
      name: 'warn-file',
      filename: 'logs/warns.log',
      level: 'warn'
    })
  ]
});
