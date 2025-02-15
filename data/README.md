# Cigarsdb filesystems persistence

## The partitions

```commandline
.
├── denormalised        # the partition after the attributes' cardinality normalisation; the partition's name indicates denormaliasiton in the SQL normal-form context  
│   └── records         # each file is a record representing the cigar's attributes
└── raw                 # the partition contains the data from the data sources
    └── records         # each file is a record representing the cigar's attributes
        ├── cigarworld  # the data from cigarworld.de
        └── noblego     # the data from noblego.de
```
