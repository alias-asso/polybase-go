package tests

// Pack name can be updated while preserving courses
// Pack courses can be updated while preserving name
// Pack courses and name can be updated simultaneously
// Updating pack to have no courses is rejected
// Updating pack with non-existent courses fails properly
// Updating pack with duplicate courses fails properly
// Updating non-existent pack returns appropriate error
// Update with no changes succeeds without modifications
