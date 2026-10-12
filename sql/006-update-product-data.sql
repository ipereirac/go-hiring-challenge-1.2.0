UPDATE products set category_id = (SELECT c.id from categories c WHERE c.code = 'Clothing')
       WHERE TRIM(code) IN('PROD001', 'PROD004', 'PROD007');

-- UPDATE products set category_id = (SELECT c.id from categories c WHERE c.code = 'Shoes')
--     WHERE code IN('PROD002', 'PROD006');
-- UPDATE products set category_id = (SELECT c.id from categories c WHERE c.code = 'Accessories')
--     WHERE code IN('PROD003', 'PROD005', 'PROD008');