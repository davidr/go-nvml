#include <nvml.h>
#include <stdlib.h>
#include <stddef.h>
#include <string.h>

// Not every function can be genericized in this way because of all the custom structs,
// but there are several nvmlGet functions we want that take a nvmlDevice_t, *char, and
// a length as arguments. These are trivial to pass as function pointers along with their,
// arguments, so we might as well save some effort.
typedef int (*gettextProperty) (nvmlDevice_t device , char *buf, unsigned int length);
int bridge_get_text_property(gettextProperty f,
                             nvmlDevice_t device,
                             char *buf,
                             unsigned int length);

// Same as above, but for integer properties
typedef int (*getintProperty) (nvmlDevice_t device , unsigned int *property);
int bridge_get_int_property(getintProperty f,
                             nvmlDevice_t device,
                             unsigned int *property);
