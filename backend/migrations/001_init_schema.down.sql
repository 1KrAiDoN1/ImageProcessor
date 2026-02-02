-- Drop triggers
DROP TRIGGER IF EXISTS update_images_updated_at ON images;
DROP TRIGGER IF EXISTS update_processing_jobs_updated_at ON processing_jobs;
DROP TRIGGER IF EXISTS update_statistics_updated_at ON statistics;
DROP TRIGGER IF EXISTS update_operation_statistics_updated_at ON operation_statistics;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS operation_statistics;
DROP TABLE IF EXISTS statistics;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS processed_images;
DROP TABLE IF EXISTS images;

