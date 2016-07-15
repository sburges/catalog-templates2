# catalog-templates

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [HSM Catalog](#hsm-catalog)
- [Catalog Templates Repository](#catalog-templates-repository)
  - [Stackato Core Services](#stackato-core-services)
  - [Other services from HSM catalog](#other-services-from-hsm-catalog)
- [Service Definition](#service-definition)

# HSM Catalog
HSM Catalog is either an s3 bucket or github repository location which is used by HSM service to present list of available services that can be used/provisioned via HSM. Each HSM service deployment can be configured, to talk to a particular HSM catalog location.

# Catalog Templates Repository
Git repository for catalog-templates holds service definitions for all the services shipped via HSM under default catalog. catalog-templates repository is source of truth for all the services shipped via HSM, except for the Stackato core services.

## Stackato Core services
For Stackato core services specifically HCF, HCE and Helion-Console catalog-templates repository has the service definition, but only the service level config file, icon, readme and version level config files are used from this repository. The SDLs for these services are provided by service teams, by copying them in the [partern-services s3 bucket](s3://helion-service-manager/partner-service)

## Other services from HSM catalog
Other services shipped via HSM catalog like all the dev grade services (dev-mysql, dev-mongo, dev-redis, dev-rabbit), RDS and Haven on Demand SDLs are also copied at the same location in service specific folders. SDLs for these services are copied to this s3 location by [CSM nightly pipeline](https://concourse.helion.lol/pipelines/hsm-nightly?groups=CSM), after it builds all the CSM extension images. These SDLs are available in catalog-templates repository, the update scripts modifies the SDLs in the repository with appropriate SDL version (based on build number) and image tags (also based on build number) and then they are copied at this s3 locations.


# Service Definition
The service definitions should follow the strict directory structure and the files should be placed in the correct locations for HSM to be able to read and interpret the service details correctly.
```
service-name/          # this should be exactly identical to the name in the SDL
├── config.json        # top level service config stores description, categories, labels etc
├── README.md
└── icon.png
├── <Product Version>   # Product Version directory, any string is supported here
│   └── <SDL Version>   # SDL Version directory, this needs be a valid semver format
│       ├── config.json   # defines upgrade configuration for the SDL version
│       └── sdl
│           └── sdl.json  # service sdl which defines the service deployment details.
```
