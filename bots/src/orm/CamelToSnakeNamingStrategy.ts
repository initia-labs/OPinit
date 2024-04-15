import { DefaultNamingStrategy, NamingStrategyInterface } from 'typeorm'
import { snakeCase } from 'lodash'

class CamelToSnakeNamingStrategy
  extends DefaultNamingStrategy
  implements NamingStrategyInterface
{
  tableName(targetName, userSpecifiedName) {
    return userSpecifiedName ? userSpecifiedName : snakeCase(targetName)
  }
  columnName(propertyName, customName, embeddedPrefixes) {
    return snakeCase(
      embeddedPrefixes.concat(customName ? customName : propertyName).join('_')
    )
  }
  columnNameCustomized(customName) {
    return customName
  }
  relationName(propertyName) {
    return snakeCase(propertyName)
  }
}

export default CamelToSnakeNamingStrategy
