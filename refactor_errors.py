import os
import re

usecase_dir = "internal/usecase"
handler_dir = "internal/delivery/http/handler"

error_map = {
    # Not Found
    "order not found": "ErrNotFound",
    "payment not found": "ErrNotFound",
    "product not found": "ErrNotFound",
    "cart item not found": "ErrNotFound",
    "shop not found": "ErrNotFound",
    
    # Forbidden
    "unauthorized to update order status": "ErrForbidden",
    "unauthorized to cancel this order": "ErrForbidden",
    "unauthorized to update fulfillment status": "ErrForbidden",
    "unauthorized to complete this order": "ErrForbidden",
    "order not found or unauthorized": "ErrNotFound", # Map to NotFound usually better if we don't want to leak existence, but let's stick to NotFound
    "unauthorized to delete this review": "ErrForbidden",
    "user has no completed order for this product": "ErrForbidden",
    "unauthorized to update this shop": "ErrForbidden",
    "shop not found for this seller": "ErrForbidden",
    "unauthorized to update this product": "ErrForbidden",
    "unauthorized to delete this product": "ErrForbidden",
    "unauthorized": "ErrForbidden",
    
    # Bad Request
    "only pending orders can be cancelled": "ErrBadRequest",
    "cart is empty": "ErrBadRequest",
    "insufficient stock": "ErrBadRequest",
    "selected cart items are invalid or already checked out": "ErrBadRequest",
    "order is no longer pending": "ErrBadRequest",
    "product not found or unavailable": "ErrBadRequest",
    "one or more categories not found": "ErrBadRequest",

    # Conflict
    "user already has a shop": "ErrConflict",
    "failed to create review": "ErrConflict",
    "user already exists": "ErrConflict",
    
    # Unauthorized
    "invalid credentials": "ErrUnauthorized",
}

def get_error_type(msg):
    for key, val in error_map.items():
        if key in msg:
            return val
    return None

# 1. Update UseCases
for filename in os.listdir(usecase_dir):
    if not filename.endswith(".go"):
        continue
    filepath = os.path.join(usecase_dir, filename)
    with open(filepath, "r") as f:
        content = f.read()
    
    # We need to replace errors.New("...") with fmt.Errorf("%w: ...", apperror.ErrX, "...")
    # Find all errors.New("...")
    matches = re.finditer(r'errors\.New\("([^"]+)"\)', content)
    replacements = []
    has_apperror = False
    has_fmt = False
    
    for match in matches:
        msg = match.group(1)
        err_type = get_error_type(msg)
        if err_type:
            new_str = f'fmt.Errorf("%w: %s", apperror.{err_type}, "{msg}")'
            replacements.append((match.group(0), new_str))
            has_apperror = True
            has_fmt = True
            
    # Add dynamic ones like errors.New("insufficient stock for " + ...)
    dyn_matches = re.finditer(r'errors\.New\("([^"]+)" \+ ([^\)]+)\)', content)
    for match in dyn_matches:
        msg = match.group(1)
        var = match.group(2)
        err_type = get_error_type(msg)
        if err_type:
            new_str = f'fmt.Errorf("%w: %s%s", apperror.{err_type}, "{msg}", {var})'
            replacements.append((match.group(0), new_str))
            has_apperror = True
            has_fmt = True
            
    if not replacements:
        continue
        
    for old, new in replacements:
        content = content.replace(old, new)
        
    if has_apperror and '"github.com/tudemaha/marketplace-be/pkg/apperror"' not in content:
        content = content.replace('"errors"\n', '"errors"\n\t"fmt"\n\t"github.com/tudemaha/marketplace-be/pkg/apperror"\n')
    elif has_fmt and '"fmt"' not in content:
        content = content.replace('"errors"\n', '"errors"\n\t"fmt"\n')
        
    with open(filepath, "w") as f:
        f.write(content)

# 2. Update Handlers
handler_replacements = {
    'err.Error()[:12] == "unauthorized"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error()[:23] == "failed to create review"': 'errors.Is(err, apperror.ErrConflict)',
    'len(err.Error()) > 18 && err.Error()[:18] == "insufficient stock"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "unauthorized"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "unauthorized to update order status"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "order not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "only pending orders can be cancelled"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "order not found or unauthorized"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "order is no longer pending"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "payment not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "product not found or unavailable" || err.Error() == "insufficient stock"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "cart item not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "insufficient stock"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "user already has a shop"': 'errors.Is(err, apperror.ErrConflict)',
    'err.Error() == "shop not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "unauthorized to update this shop"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "shop not found for this seller"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "one or more categories not found"': 'errors.Is(err, apperror.ErrBadRequest)',
    'err.Error() == "product not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "unauthorized to update this product"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "unauthorized to delete this product"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "user has no completed order for this product"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "unauthorized to delete this review"': 'errors.Is(err, apperror.ErrForbidden)',
    'err.Error() == "review not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "order not found or unauthorized" || err.Error() == "payment not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "payment not found" || err.Error() == "order not found"': 'errors.Is(err, apperror.ErrNotFound)',
    'err.Error() == "cart is empty" || err.Error() == "insufficient stock for one or more items during checkout"': 'errors.Is(err, apperror.ErrBadRequest)',
    'len(err.Error()) > 18 && err.Error()[:18] == "insufficient stock"': 'errors.Is(err, apperror.ErrBadRequest)',
}

for filename in os.listdir(handler_dir):
    if not filename.endswith(".go"):
        continue
    filepath = os.path.join(handler_dir, filename)
    with open(filepath, "r") as f:
        content = f.read()
        
    has_replaced = False
    for old, new in handler_replacements.items():
        if old in content:
            content = content.replace(old, new)
            has_replaced = True
            
    # Auth handler manual ones
    if "auth_handler.go" in filename:
        content = content.replace('err.Error() == "invalid credentials"', 'errors.Is(err, apperror.ErrUnauthorized)')
        content = content.replace('err.Error() == "user already exists"', 'errors.Is(err, apperror.ErrConflict)')
        has_replaced = True
            
    if has_replaced:
        if '"errors"' not in content:
            content = content.replace('"net/http"\n', '"errors"\n\t"net/http"\n')
        if '"github.com/tudemaha/marketplace-be/pkg/apperror"' not in content:
            content = content.replace('"github.com/tudemaha/marketplace-be/pkg/response"', '"github.com/tudemaha/marketplace-be/pkg/apperror"\n\t"github.com/tudemaha/marketplace-be/pkg/response"')
            
    with open(filepath, "w") as f:
        f.write(content)

