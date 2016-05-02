
#include "nvmlbridge.h"
#include <stdlib.h>

// Not every function can be genericized in this way because of all the custom structs,
// but there are several nvmlGet functions we want that take a nvmlDevice_t, *char, and
// a length as arguments. These are trivial to pass as function pointers along with their,
// arguments, so we might as well save some effort.
//
int bridge_get_text_property(gettextProperty f,
                             nvmlDevice_t device,
                             char *buf,
                             unsigned int length)
{
    nvmlReturn_t ret;

    ret = f(device, buf, length);

    if (ret == NVML_SUCCESS) {
        return(EXIT_SUCCESS);
    } else {
        return(EXIT_FAILURE);
    }
}

// Same as above, but for integer properties
int bridge_get_int_property(getintProperty f,
                             nvmlDevice_t device,
                             unsigned int *property)
{
    nvmlReturn_t ret;

    ret = f(device, property);

    if (ret == NVML_SUCCESS) {
        return(EXIT_SUCCESS);
    } else {
        return(EXIT_FAILURE);
    }
}

