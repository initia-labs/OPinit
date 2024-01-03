/* eslint-disable no-undef */
const { DefaultNamingStrategy } = require('typeorm');
const { snakeCase } = require('lodash');

class CamelToSnakeNamingStrategy extends DefaultNamingStrategy {
  tableName(targetName, userSpecifiedName) {
    return userSpecifiedName ? userSpecifiedName : snakeCase(targetName);
  }
  columnName(propertyName, customName, embeddedPrefixes) {
    return snakeCase(embeddedPrefixes.concat(customName ? customName : propertyName).join('_'));
  }
  columnNameCustomized(customName) {
    return customName;
  }
  relationName(propertyName) {
    return snakeCase(propertyName);
  }
}

module.exports = CamelToSnakeNamingStrategy;