/**
 * Wasp API
 * REST API for the Wasp node
 *
 * OpenAPI spec version: 0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { HttpFile } from '../http/http';

export class BaseToken {
    /**
    * The token decimals
    */
    'decimals': number;
    /**
    * The base token name
    */
    'name': string;
    /**
    * The token subunit
    */
    'subunit': string;
    /**
    * The ticker symbol
    */
    'tickerSymbol': string;
    /**
    * The token unit
    */
    'unit': string;
    /**
    * Whether or not the token uses a metric prefix
    */
    'useMetricPrefix': boolean;

    static readonly discriminator: string | undefined = undefined;

    static readonly attributeTypeMap: Array<{name: string, baseName: string, type: string, format: string}> = [
        {
            "name": "decimals",
            "baseName": "decimals",
            "type": "number",
            "format": "int32"
        },
        {
            "name": "name",
            "baseName": "name",
            "type": "string",
            "format": "string"
        },
        {
            "name": "subunit",
            "baseName": "subunit",
            "type": "string",
            "format": "string"
        },
        {
            "name": "tickerSymbol",
            "baseName": "tickerSymbol",
            "type": "string",
            "format": "string"
        },
        {
            "name": "unit",
            "baseName": "unit",
            "type": "string",
            "format": "string"
        },
        {
            "name": "useMetricPrefix",
            "baseName": "useMetricPrefix",
            "type": "boolean",
            "format": "boolean"
        }    ];

    static getAttributeTypeMap() {
        return BaseToken.attributeTypeMap;
    }

    public constructor() {
    }
}
