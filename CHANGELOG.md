## v1.11.0 (unreleased)
ENHANCEMENTS:
- `azapi_resource_action` resource: Support `HEAD` method.


## v1.10.0
ENHANCEMENTS:
- `azapi_resource` data source: When creating `Microsoft.Resources/subscriptions`, `resource_id` is optional and defaults to the ID of the default subscription.
- Add a new logger to record the traffic in a structured way.
- `azapi`: Support `endpoint` block, which is used to configure the endpoints of the Azure Clouds.
- `azapi_resource_action` resource: Support `GET` method.
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/505b813ce50368156e3da1b86f07977b5a913be9

BUG FIXES:
- Fix a bug that `body` is not set when import with an unrecognized `api-version`.
- Fix a bug that deploy time constants are not removed from the request body when using `azapi_update_resource` resource.

## v1.9.0
FEATURES:
- **New Data Source**: azapi_resource_list
- **New Data Source**: azapi_resource_id

ENHANCEMENTS:
- `azapi_resource` resource/data source: When creating `Microsoft.Resources/resourceGroups`, `parent_id` is optional and defaults to the ID of the default subscription.
- `azapi_resource` resource: Support `ignore_body_changes` field, which is used to ignore some properties when comparing the resource with its current state.
- `azapi_update_resource` resource: Support `ignore_body_changes` field, which is used to ignore some properties when comparing the resource with its current state.
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/1d8fec8184258cdf967b1288b156e01f7cbc8ca9

BUG FIXES:
- Fix a bug that `azapi_resource` resource doesn't store the `id` in the state when error happens during the creation.
- Fix a bug that errors from the polling API which don't follow the ARM LRO guidelines are not handled properly.

## v1.8.0
FEATURES:

ENHANCEMENTS:
- `azapi_resource_action`: Support provider action.
- Update bicep types to https://ms-henglu/bicep-types-az/commit/c616eb1ad4980f63c0d6b436a63701e175a62224

BUG FIXES:
- Fix a bug that resource id for type `Microsoft.Resources/providers` is not parsed correctly.
- Fix a bug that resource id for type `Microsoft.Resources/tenants` is not parsed correctly.

## v1.7.0
FEATURES:
- **New Resource**: azapi_data_plane_resource
- `azapi`: Support `use_msi` and `use_cli` features.
- `azapi`: Support `auxiliary_tenant_ids` field, which is required for multi-tenancy and cross-tenant scenarios.
- `azapi`: Support `custom_correlation_request_id` field, which is used to specify the correlation request id.

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/0536b68e779fba100b9fbe32737c38d75396e2cf

BUG FIXES:
- Fix a bug that provider crashes when loading azure schema.

## v1.6.0
FEATURES:

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/da15d0376faa02a6e891dee315910535cef2a13f

BUG FIXES:
- Fix the bug that the headers are not stored in the log.

## v1.5.0
FEATURES:

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/b8626aecc5f47b70086580956adfcd1e267a49e6

BUG FIXES:

## v1.4.0
FEATURES:
- `azapi`: Support `default_name`, `default_naming_prefix` and `defualt_naming_suffix` features.

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/a915acab5788d890aed774ec818535b44311d16d

BUG FIXES:
- Fix a bug that when apply failed there still are some attributes stored in the state.
- Fix a bug that schema validation requires redundant `name` fields both in resource and in body.

## v1.3.0
FEATURES:
- `azapi`: Support OIDC authentication.

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/78ec1b99699a4bf44869bd13f1b0ed7d92a99c27
- `azapi_resource`: `ignore_missing_property` will also ignore the sensitive properties returned in asterisks.

BUG FIXES:
- Fix a document typo.

## v1.2.0
FEATURES:
- `azapi`: Support `client_certificate_password` option.

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/019b2d62fe84508582b8c54ce3d91c2b4840e624

BUG FIXES:

## v1.1.0
FEATURES:

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/e641570bedc5004498d3e374adb60fdfd3521b09

BUG FIXES:
- `azapi_resource_action`: The `output` is not refreshed when `body` is changed.

## v1.0.0
FEATURES:

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/a6dabb0cd645c17a1accf3ec1be4d7930e982b23

BUG FIXES:

BREAKING CHANGES:
- `azapi_resource`: `ignore_missing_property`'s default value changed from `false` to `true`.
- `azapi_update_resource`: `ignore_missing_property`'s default value changed from `false` to `true`.

## v0.6.0
FEATURES:

ENHANCEMENTS:
- `azapi_resource_action`: Supports `locks` which used to prevent modifying resources at the same time.
- `azapi_resource_action`: Supports parse response which `Content-Type` is `text/plain`.
- Improve validation on `type`, `parent_id` and `resource_id`.
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/5451fcd5e1bf4d8313d475d8e3dc28efc7a77e2a

BUG FIXES:

## v0.5.0
FEATURES:
- **New Data Source**: azapi_resource_action
- **New Resource**: azapi_resource_action

ENHANCEMENTS:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/813d8bbc9ecf432a2a0ff2769627592fae34369f

BUG FIXES:
- DefaultAzureCredential authentication failed because empty clientId is set

## v0.4.0
FEATURES:

ENHANCEMENTS:
- `azapi_resource`: Supports default api-version when importing existing resource into terraform state.
- `azapi_resource`: Supports `locks` which used to prevent creating/modifying/deleting resources at the same time.
- `azapi_update_resource`: Supports `locks` which used to prevent creating/modifying/deleting resources at the same time.
- `azapi_resource` data source: Supports configuring `resource_id` as an alternative way to configure `name` and `parent_id`.
- `azapi` provider: Supports `partner_id`, `disable_terraform_partner_id` and `disable_terraform_partner_id`.
- Update bicep types to https://github.com/Azure/bicep-types-az/commit/ea703e2aba0d1c024f33124ee2cd34bc0c6084b5

BUG FIXES:

## v0.3.0
FEATURES:

ENHANCEMENTS:
- Update bicep types to https://github.com/Azure/bicep-types-az/commit/644ff521c92ce8d493f6da977af12377f32abffc

BUG FIXES:

## v0.2.1
FEATURES:

ENHANCEMENTS:

BUG FIXES:
- Improve error message for schema validation failure.
- DefaultAzureCredential reads the client ID of a user-assigned managed identity.
- Fix the modification is not working, when use `azapi_update_resource` to modify additional properties.
- Fix crash when use `azapi_update_resource` on a resource whose id is null
- Fix crash when the discriminated type is not in the embedded schema

## v0.2.0
FEATURES:

ENHANCEMENTS:
- Setting `response_export_values = ["*"]` will export the full response body.
- Update bicep types to https://github.com/Azure/bicep-types-az/commit/57f3ecc750648562cf170ef456ef39533872b101

BUG FIXES:
- Fix incorrect ID format in the imported `azapi_resource` resource. 
- Fix incorrect `body` content in the imported `azapi_resource` resource.

## v0.1.1

FEATURES:

ENHANCEMENTS:

BUG FIXES:

- Fix document format.

## v1.1.0

FEATURES:
- **New Data Source**: azapi_resource
- **New Resource**: azapi_resource
- **New Resource**: azapi_update_resource
- **Provider feature**: support default location and default tags

ENHANCEMENTS:

BUG FIXES:
