
#include "nvmlbridge.h"
#include <stdlib.h>

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

